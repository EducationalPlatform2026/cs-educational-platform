#!/usr/bin/env bash
# Seed script — populates the CS Educational Platform with realistic test data.
# Usage: bash scripts/seed.sh
# Requires: curl, jq, docker (for the admin insert)
set -euo pipefail

API="http://localhost:8080"
DB_CONTAINER="csplatform_db"
DB_NAME="csplatform_db"
DB_USER="csplatform"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
log()  { echo -e "${GREEN}[seed]${NC} $*"; }
warn() { echo -e "${YELLOW}[warn]${NC} $*"; }

# ── helpers ──────────────────────────────────────────────────────────────────

register() {          # register email pass first last role → token
  local email=$1 pass=$2 first=$3 last=$4 role=$5
  local resp
  resp=$(curl -sf -X POST "$API/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$pass\",\"first_name\":\"$first\",\"last_name\":\"$last\",\"role\":\"$role\"}" \
    2>&1) || { warn "register $email: $resp"; echo ""; return 0; }
  echo "$resp" | jq -r '.token // ""'
}

login() {             # login email pass → token
  local email=$1 pass=$2
  curl -sf -X POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$pass\"}" | jq -r '.token'
}

create_course() {     # token title desc published → course_id
  local tok=$1 title=$2 desc=$3 pub=$4
  curl -sf -X POST "$API/courses" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $tok" \
    -d "{\"title\":\"$title\",\"description\":\"$desc\",\"is_published\":$pub}" | jq -r '.id'
}

create_exercise() {   # tok course_id title desc instructions diff lang template_code time_ms mem_kb pub → ex_id
  local tok=$1 cid=$2 title=$3 desc=$4 instr=$5 diff=$6 lang=$7 tmpl=$8 time=$9 mem=${10} pub=${11}
  curl -sf -X POST "$API/courses/$cid/exercises" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $tok" \
    -d "{\"title\":\"$title\",\"description\":\"$desc\",\"instructions\":\"$instr\",\"difficulty\":\"$diff\",\"language\":\"$lang\",\"template_code\":\"$tmpl\",\"time_limit_ms\":$time,\"memory_limit_kb\":$mem,\"is_published\":$pub}" | jq -r '.id'
}

add_test_case() {     # tok ex_id input expected hidden ordinal
  local tok=$1 eid=$2 inp=$3 exp=$4 hidden=$5 ord=$6
  local inp_json exp_json
  inp_json=$(jq -Rs . <<<"$inp")   # safely quote multi-line strings
  exp_json=$(jq -Rs . <<<"$exp")
  curl -sf -X POST "$API/exercises/$eid/test-cases" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $tok" \
    -d "{\"input\":$inp_json,\"expected_output\":$exp_json,\"is_hidden\":$hidden,\"ordinal\":$ord}" > /dev/null
}

enroll() {            # student_tok course_id
  curl -sf -X POST "$API/courses/$2/enroll" \
    -H "Authorization: Bearer $1" \
    -H "Content-Type: application/json" > /dev/null || true
}

submit() {            # student_tok ex_id lang code
  local tok=$1 eid=$2 lang=$3 code=$4
  local code_json
  code_json=$(jq -Rs . <<<"$code")
  curl -sf -X POST "$API/exercises/$eid/submit" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $tok" \
    -d "{\"language\":\"$lang\",\"code\":$code_json}" > /dev/null || true
}

# ── 1. Create users ───────────────────────────────────────────────────────────

log "Creating professors…"
TOK_PROF1=$(register "alice.morgan@cs.edu"   "Password123" "Alice"   "Morgan"   "professor")
TOK_PROF2=$(register "bob.carter@cs.edu"     "Password123" "Bob"     "Carter"   "professor")
TOK_PROF3=$(register "carol.james@cs.edu"    "Password123" "Carol"   "James"    "professor")

log "Creating teaching assistants…"
TOK_TA1=$(register "david.kim@cs.edu"        "Password123" "David"   "Kim"      "teaching_assistant")
TOK_TA2=$(register "emma.silva@cs.edu"       "Password123" "Emma"    "Silva"    "teaching_assistant")
TOK_TA3=$(register "frank.liu@cs.edu"        "Password123" "Frank"   "Liu"      "teaching_assistant")
TOK_TA4=$(register "grace.patel@cs.edu"      "Password123" "Grace"   "Patel"    "teaching_assistant")

log "Creating students…"
TOK_S1=$(register  "henry.walsh@student.edu" "Password123" "Henry"   "Walsh"    "student")
TOK_S2=$(register  "isabella.chen@student.edu" "Password123" "Isabella" "Chen"  "student")
TOK_S3=$(register  "jack.novak@student.edu"  "Password123" "Jack"    "Novak"    "student")
TOK_S4=$(register  "kate.torres@student.edu" "Password123" "Kate"    "Torres"   "student")
TOK_S5=$(register  "liam.oconnor@student.edu" "Password123" "Liam"   "OConnor"  "student")
TOK_S6=$(register  "maya.singh@student.edu"  "Password123" "Maya"    "Singh"    "student")
TOK_S7=$(register  "noah.berg@student.edu"   "Password123" "Noah"    "Berg"     "student")
TOK_S8=$(register  "olivia.martin@student.edu" "Password123" "Olivia" "Martin"  "student")

log "Creating admin user via direct DB insert…"
docker exec "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c \
  "INSERT INTO users (email, password_hash, first_name, last_name, role)
   VALUES ('admin@cs.edu', '\$2a\$12\$mB1K3fQIWD.VECVFkqGSsuT.TqhvU1s2ILt8S3HaqPTt28GE6V.bq', 'Admin', 'User', 'admin')
   ON CONFLICT (email) DO NOTHING;" > /dev/null
# password_hash above = bcrypt("Password123", cost=12)
log "Admin: admin@cs.edu / Password123"

# Re-login for tokens that may have been skipped (already registered)
[ -z "$TOK_PROF1" ] && TOK_PROF1=$(login "alice.morgan@cs.edu" "Password123")
[ -z "$TOK_PROF2" ] && TOK_PROF2=$(login "bob.carter@cs.edu" "Password123")
[ -z "$TOK_PROF3" ] && TOK_PROF3=$(login "carol.james@cs.edu" "Password123")
[ -z "$TOK_TA1" ]   && TOK_TA1=$(login "david.kim@cs.edu" "Password123")
[ -z "$TOK_TA2" ]   && TOK_TA2=$(login "emma.silva@cs.edu" "Password123")
[ -z "$TOK_TA3" ]   && TOK_TA3=$(login "frank.liu@cs.edu" "Password123")
[ -z "$TOK_TA4" ]   && TOK_TA4=$(login "grace.patel@cs.edu" "Password123")
[ -z "$TOK_S1" ]    && TOK_S1=$(login "henry.walsh@student.edu" "Password123")
[ -z "$TOK_S2" ]    && TOK_S2=$(login "isabella.chen@student.edu" "Password123")
[ -z "$TOK_S3" ]    && TOK_S3=$(login "jack.novak@student.edu" "Password123")
[ -z "$TOK_S4" ]    && TOK_S4=$(login "kate.torres@student.edu" "Password123")
[ -z "$TOK_S5" ]    && TOK_S5=$(login "liam.oconnor@student.edu" "Password123")
[ -z "$TOK_S6" ]    && TOK_S6=$(login "maya.singh@student.edu" "Password123")
[ -z "$TOK_S7" ]    && TOK_S7=$(login "noah.berg@student.edu" "Password123")
[ -z "$TOK_S8" ]    && TOK_S8=$(login "olivia.martin@student.edu" "Password123")

log "All users ready."

# ── 2. Create courses ─────────────────────────────────────────────────────────

log "Creating courses…"

CID_PYTHON=$(create_course "$TOK_PROF1" \
  "Introduction to Python Programming" \
  "A beginner-friendly course covering Python fundamentals: variables, loops, functions, and basic data structures." \
  true)

CID_ALGO=$(create_course "$TOK_PROF1" \
  "Data Structures & Algorithms" \
  "Deep dive into arrays, linked lists, trees, graphs, sorting, and searching. Prepares students for technical interviews." \
  true)

CID_GO=$(create_course "$TOK_PROF2" \
  "Systems Programming in Go" \
  "Concurrent and systems-level programming using Go. Topics: goroutines, channels, interfaces, and the standard library." \
  true)

CID_WEB=$(create_course "$TOK_PROF3" \
  "Web Development Fundamentals" \
  "Full-stack basics with HTML, CSS, JavaScript. Students build and deploy small web applications." \
  true)

CID_CPP=$(create_course "$TOK_PROF2" \
  "C++ for Competitive Programming" \
  "Unpublished draft: STL, templates, and advanced problem-solving techniques for competitive programming." \
  false)

log "Courses created."

# ── 3. Enroll students and TAs ────────────────────────────────────────────────

log "Enrolling members…"

# Python course — large enrollment
for tok in $TOK_S1 $TOK_S2 $TOK_S3 $TOK_S4 $TOK_S5 $TOK_S6 $TOK_S7 $TOK_S8; do
  enroll "$tok" "$CID_PYTHON"
done
enroll "$TOK_TA1" "$CID_PYTHON"
enroll "$TOK_TA2" "$CID_PYTHON"

# Algo course
for tok in $TOK_S1 $TOK_S2 $TOK_S3 $TOK_S5 $TOK_S7; do
  enroll "$tok" "$CID_ALGO"
done
enroll "$TOK_TA1" "$CID_ALGO"
enroll "$TOK_TA3" "$CID_ALGO"

# Go course
for tok in $TOK_S2 $TOK_S4 $TOK_S6 $TOK_S8; do
  enroll "$tok" "$CID_GO"
done
enroll "$TOK_TA4" "$CID_GO"

# Web course
for tok in $TOK_S3 $TOK_S5 $TOK_S6 $TOK_S7 $TOK_S8; do
  enroll "$tok" "$CID_WEB"
done
enroll "$TOK_TA2" "$CID_WEB"
enroll "$TOK_TA3" "$CID_WEB"

log "Enrollment done."

# ── 4. Python exercises ───────────────────────────────────────────────────────

log "Creating Python exercises…"

EX_HELLO=$(create_exercise "$TOK_PROF1" "$CID_PYTHON" \
  "Hello, World!" \
  "Your very first Python program." \
  "Write a Python program that prints exactly: Hello, World!" \
  "easy" "python" \
  'print("Hello, World!")' \
  2000 65536 true)

add_test_case "$TOK_PROF1" "$EX_HELLO" "" "Hello, World!" false 1
add_test_case "$TOK_PROF1" "$EX_HELLO" "" "Hello, World!" true  2

EX_SUM=$(create_exercise "$TOK_PROF1" "$CID_PYTHON" \
  "Sum of Two Numbers" \
  "Read two integers from stdin and print their sum." \
  "Read two integers from stdin (one per line) and print their sum on a single line." \
  "easy" "python" \
  'a = int(input())
b = int(input())
print(a + b)' \
  2000 65536 true)

add_test_case "$TOK_PROF1" "$EX_SUM" "3
5" "8" false 1
add_test_case "$TOK_PROF1" "$EX_SUM" "0
0" "0" false 2
add_test_case "$TOK_PROF1" "$EX_SUM" "-7
4" "-3" true  3
add_test_case "$TOK_PROF1" "$EX_SUM" "1000000
999999" "1999999" true 4

EX_FIZ=$(create_exercise "$TOK_PROF1" "$CID_PYTHON" \
  "FizzBuzz" \
  "The classic FizzBuzz problem." \
  'Read an integer N from stdin. Print numbers from 1 to N (inclusive), one per line. Replace multiples of 3 with "Fizz", multiples of 5 with "Buzz", and multiples of both with "FizzBuzz".' \
  "easy" "python" \
  'n = int(input())
for i in range(1, n + 1):
    if i % 15 == 0:
        print("FizzBuzz")
    elif i % 3 == 0:
        print("Fizz")
    elif i % 5 == 0:
        print("Buzz")
    else:
        print(i)' \
  2000 65536 true)

add_test_case "$TOK_PROF1" "$EX_FIZ" "15" "1
2
Fizz
4
Buzz
Fizz
7
8
Fizz
Buzz
11
Fizz
13
14
FizzBuzz" false 1
add_test_case "$TOK_PROF1" "$EX_FIZ" "5" "1
2
Fizz
4
Buzz" false 2
add_test_case "$TOK_PROF1" "$EX_FIZ" "1" "1" true 3

EX_PALINDROME=$(create_exercise "$TOK_PROF1" "$CID_PYTHON" \
  "Palindrome Check" \
  "Determine whether a string is a palindrome." \
  'Read a single line of text. Print "YES" if it is a palindrome (case-insensitive, ignore spaces), "NO" otherwise.' \
  "medium" "python" \
  's = input().replace(" ", "").lower()
print("YES" if s == s[::-1] else "NO")' \
  2000 65536 true)

add_test_case "$TOK_PROF1" "$EX_PALINDROME" "racecar" "YES" false 1
add_test_case "$TOK_PROF1" "$EX_PALINDROME" "hello" "NO" false 2
add_test_case "$TOK_PROF1" "$EX_PALINDROME" "A man a plan a canal Panama" "YES" true 3
add_test_case "$TOK_PROF1" "$EX_PALINDROME" "Was it a car or a cat I saw" "YES" true 4

EX_PRIMES=$(create_exercise "$TOK_PROF1" "$CID_PYTHON" \
  "Sieve of Eratosthenes" \
  "Find all prime numbers up to N." \
  "Read an integer N from stdin. Print all prime numbers up to N (inclusive), one per line, in ascending order." \
  "medium" "python" \
  'n = int(input())
sieve = [True] * (n + 1)
sieve[0] = sieve[1] = False
for i in range(2, int(n**0.5) + 1):
    if sieve[i]:
        for j in range(i*i, n+1, i):
            sieve[j] = False
for i in range(2, n+1):
    if sieve[i]:
        print(i)' \
  3000 65536 true)

add_test_case "$TOK_PROF1" "$EX_PRIMES" "10" "2
3
5
7" false 1
add_test_case "$TOK_PROF1" "$EX_PRIMES" "2" "2" false 2
add_test_case "$TOK_PROF1" "$EX_PRIMES" "1" "" true 3
add_test_case "$TOK_PROF1" "$EX_PRIMES" "50" "2
3
5
7
11
13
17
19
23
29
31
37
41
43
47" true 4

EX_FIBO=$(create_exercise "$TOK_PROF1" "$CID_PYTHON" \
  "Fibonacci Sequence" \
  "Generate the first N Fibonacci numbers." \
  "Read an integer N from stdin. Print the first N Fibonacci numbers separated by spaces on one line. F(1)=1, F(2)=1." \
  "easy" "python" \
  'n = int(input())
a, b = 1, 1
result = []
for _ in range(n):
    result.append(str(a))
    a, b = b, a + b
print(" ".join(result))' \
  2000 65536 true)

add_test_case "$TOK_PROF1" "$EX_FIBO" "1" "1" false 1
add_test_case "$TOK_PROF1" "$EX_FIBO" "5" "1 1 2 3 5" false 2
add_test_case "$TOK_PROF1" "$EX_FIBO" "10" "1 1 2 3 5 8 13 21 34 55" true 3

# ── 5. Algorithms exercises ───────────────────────────────────────────────────

log "Creating Algorithms exercises…"

EX_BSEARCH=$(create_exercise "$TOK_PROF1" "$CID_ALGO" \
  "Binary Search" \
  "Implement binary search on a sorted array." \
  'First line: N (size of array). Second line: N space-separated integers in ascending order. Third line: target integer. Print the 0-based index where target is found, or -1 if not present.' \
  "easy" "python" \
  'n = int(input())
arr = list(map(int, input().split()))
target = int(input())
lo, hi = 0, n - 1
result = -1
while lo <= hi:
    mid = (lo + hi) // 2
    if arr[mid] == target:
        result = mid
        break
    elif arr[mid] < target:
        lo = mid + 1
    else:
        hi = mid - 1
print(result)' \
  2000 65536 true)

add_test_case "$TOK_PROF1" "$EX_BSEARCH" "5
1 3 5 7 9
7" "3" false 1
add_test_case "$TOK_PROF1" "$EX_BSEARCH" "5
1 3 5 7 9
4" "-1" false 2
add_test_case "$TOK_PROF1" "$EX_BSEARCH" "1
42
42" "0" true 3

EX_SORT=$(create_exercise "$TOK_PROF1" "$CID_ALGO" \
  "Merge Sort" \
  "Implement merge sort and sort an array of integers." \
  "First line: N. Second line: N space-separated integers. Print the sorted array on one line, space-separated." \
  "medium" "python" \
  'def merge_sort(arr):
    if len(arr) <= 1:
        return arr
    mid = len(arr) // 2
    left = merge_sort(arr[:mid])
    right = merge_sort(arr[mid:])
    result = []
    i = j = 0
    while i < len(left) and j < len(right):
        if left[i] <= right[j]:
            result.append(left[i]); i += 1
        else:
            result.append(right[j]); j += 1
    return result + left[i:] + right[j:]

n = int(input())
arr = list(map(int, input().split()))
print(*merge_sort(arr))' \
  3000 65536 true)

add_test_case "$TOK_PROF1" "$EX_SORT" "5
3 1 4 1 5" "1 1 3 4 5" false 1
add_test_case "$TOK_PROF1" "$EX_SORT" "1
7" "7" false 2
add_test_case "$TOK_PROF1" "$EX_SORT" "6
-3 0 5 -1 2 8" "-3 -1 0 2 5 8" true 3

EX_STACK=$(create_exercise "$TOK_PROF1" "$CID_ALGO" \
  "Valid Parentheses" \
  "Check if a string of brackets is valid using a stack." \
  'Read a string containing only "(", ")", "{", "}", "[", "]". Print "VALID" if brackets are properly nested and closed, "INVALID" otherwise.' \
  "medium" "python" \
  's = input()
stack = []
pairs = {")": "(", "}": "{", "]": "["}
for c in s:
    if c in "({[":
        stack.append(c)
    elif c in pairs:
        if not stack or stack[-1] != pairs[c]:
            print("INVALID")
            exit()
        stack.pop()
print("VALID" if not stack else "INVALID")' \
  2000 65536 true)

add_test_case "$TOK_PROF1" "$EX_STACK" "(){}[]" "VALID" false 1
add_test_case "$TOK_PROF1" "$EX_STACK" "([)]" "INVALID" false 2
add_test_case "$TOK_PROF1" "$EX_STACK" "{[]}" "VALID" true 3
add_test_case "$TOK_PROF1" "$EX_STACK" "(((" "INVALID" true 4

EX_BFS=$(create_exercise "$TOK_PROF1" "$CID_ALGO" \
  "BFS Shortest Path" \
  "Find the shortest path in an unweighted graph using BFS." \
  'First line: N M (nodes 1..N, M edges). Next M lines: u v (undirected edge). Last line: S T (source and target). Print shortest path length, or -1 if unreachable.' \
  "hard" "python" \
  'from collections import deque
n, m = map(int, input().split())
adj = [[] for _ in range(n + 1)]
for _ in range(m):
    u, v = map(int, input().split())
    adj[u].append(v)
    adj[v].append(u)
s, t = map(int, input().split())
dist = [-1] * (n + 1)
dist[s] = 0
q = deque([s])
while q:
    u = q.popleft()
    for v in adj[u]:
        if dist[v] == -1:
            dist[v] = dist[u] + 1
            q.append(v)
print(dist[t])' \
  3000 131072 true)

add_test_case "$TOK_PROF1" "$EX_BFS" "4 4
1 2
2 3
3 4
1 3
1 4" "2" false 1
add_test_case "$TOK_PROF1" "$EX_BFS" "3 1
1 2
1 3" "-1" false 2
add_test_case "$TOK_PROF1" "$EX_BFS" "2 1
1 2
1 2" "1" true 3

# ── 6. Go exercises ───────────────────────────────────────────────────────────

log "Creating Go exercises…"

EX_GO_HELLO=$(create_exercise "$TOK_PROF2" "$CID_GO" \
  "Hello, Gopher!" \
  "Your first Go program." \
  'Write a Go program that prints: Hello, Gopher!' \
  "easy" "go" \
  'package main

import "fmt"

func main() {
    fmt.Println("Hello, Gopher!")
}' \
  3000 65536 true)

add_test_case "$TOK_PROF2" "$EX_GO_HELLO" "" "Hello, Gopher!" false 1
add_test_case "$TOK_PROF2" "$EX_GO_HELLO" "" "Hello, Gopher!" true  2

EX_GO_GOROUTINE=$(create_exercise "$TOK_PROF2" "$CID_GO" \
  "Concurrent Counter" \
  "Use goroutines and a WaitGroup to count to N concurrently." \
  'Read N from stdin. Launch N goroutines each printing its index (1-based) in the format "Worker N done". Use sync.WaitGroup. Order of output may vary — the judge checks that all N lines are present.' \
  "medium" "go" \
  'package main

import (
    "fmt"
    "sync"
)

func main() {
    var n int
    fmt.Scan(&n)
    var wg sync.WaitGroup
    for i := 1; i <= n; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            fmt.Printf("Worker %d done\n", id)
        }(i)
    }
    wg.Wait()
}' \
  3000 65536 true)

add_test_case "$TOK_PROF2" "$EX_GO_GOROUTINE" "3" "Worker 1 done
Worker 2 done
Worker 3 done" false 1

EX_GO_CHAN=$(create_exercise "$TOK_PROF2" "$CID_GO" \
  "Fan-out with Channels" \
  "Distribute work across goroutines using channels." \
  'Read N from stdin. Send integers 1..N to a jobs channel, have 3 worker goroutines double each value and send results to a results channel. Print all results sorted ascending on one line, space-separated.' \
  "hard" "go" \
  'package main

import (
    "fmt"
    "sort"
    "sync"
)

func worker(jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
    defer wg.Done()
    for j := range jobs {
        results <- j * 2
    }
}

func main() {
    var n int
    fmt.Scan(&n)
    jobs := make(chan int, n)
    results := make(chan int, n)
    var wg sync.WaitGroup
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go worker(jobs, results, &wg)
    }
    for i := 1; i <= n; i++ {
        jobs <- i
    }
    close(jobs)
    wg.Wait()
    close(results)
    out := []int{}
    for r := range results {
        out = append(out, r)
    }
    sort.Ints(out)
    for i, v := range out {
        if i > 0 { fmt.Print(" ") }
        fmt.Print(v)
    }
    fmt.Println()
}' \
  5000 131072 true)

add_test_case "$TOK_PROF2" "$EX_GO_CHAN" "5" "2 4 6 8 10" false 1
add_test_case "$TOK_PROF2" "$EX_GO_CHAN" "1" "2" true 2

# ── 7. Web/JavaScript exercises ───────────────────────────────────────────────

log "Creating JavaScript exercises…"

EX_JS_HELLO=$(create_exercise "$TOK_PROF3" "$CID_WEB" \
  "Hello from Node" \
  "Print a greeting from a Node.js script." \
  'Write a JavaScript program that prints: Hello from Node.js!' \
  "easy" "javascript" \
  'console.log("Hello from Node.js!");' \
  2000 65536 true)

add_test_case "$TOK_PROF3" "$EX_JS_HELLO" "" "Hello from Node.js!" false 1

EX_JS_REVERSE=$(create_exercise "$TOK_PROF3" "$CID_WEB" \
  "Reverse a String" \
  "Reverse a string using JavaScript." \
  'Read a line from stdin and print it reversed.' \
  "easy" "javascript" \
  'const readline = require("readline");
const rl = readline.createInterface({ input: process.stdin });
rl.on("line", (line) => {
    console.log(line.split("").reverse().join(""));
    rl.close();
});' \
  2000 65536 true)

add_test_case "$TOK_PROF3" "$EX_JS_REVERSE" "hello" "olleh" false 1
add_test_case "$TOK_PROF3" "$EX_JS_REVERSE" "OpenAI" "IAnepO" false 2
add_test_case "$TOK_PROF3" "$EX_JS_REVERSE" "racecar" "racecar" true 3

EX_JS_REDUCE=$(create_exercise "$TOK_PROF3" "$CID_WEB" \
  "Array Sum with reduce()" \
  "Use Array.prototype.reduce to sum an array." \
  'Read N, then N space-separated numbers. Use reduce() to compute and print their sum.' \
  "medium" "javascript" \
  'const lines = [];
require("readline").createInterface({ input: process.stdin })
  .on("line", l => lines.push(l.trim()))
  .on("close", () => {
      const nums = lines[1].split(" ").map(Number);
      console.log(nums.reduce((acc, x) => acc + x, 0));
  });' \
  2000 65536 true)

add_test_case "$TOK_PROF3" "$EX_JS_REDUCE" "4
1 2 3 4" "10" false 1
add_test_case "$TOK_PROF3" "$EX_JS_REDUCE" "3
-1 0 1" "0" true 2

# ── 8. Sample submissions ─────────────────────────────────────────────────────

log "Creating sample submissions…"

# Henry: correct FizzBuzz
submit "$TOK_S1" "$EX_FIZ" "python" 'n = int(input())
for i in range(1, n + 1):
    if i % 15 == 0:
        print("FizzBuzz")
    elif i % 3 == 0:
        print("Fizz")
    elif i % 5 == 0:
        print("Buzz")
    else:
        print(i)'

# Henry: wrong palindrome attempt
submit "$TOK_S1" "$EX_PALINDROME" "python" 's = input().lower()
print("YES" if s == s[::-1] else "NO")'

# Isabella: correct sum
submit "$TOK_S2" "$EX_SUM" "python" 'a = int(input())
b = int(input())
print(a + b)'

# Isabella: binary search
submit "$TOK_S2" "$EX_BSEARCH" "python" 'n = int(input())
arr = list(map(int, input().split()))
target = int(input())
print(arr.index(target) if target in arr else -1)'

# Jack: hello world Go
submit "$TOK_S3" "$EX_GO_HELLO" "go" 'package main
import "fmt"
func main() { fmt.Println("Hello, Gopher!") }'

# Kate: Fibonacci
submit "$TOK_S4" "$EX_FIBO" "python" 'n = int(input())
a, b = 1, 1
res = []
for _ in range(n):
    res.append(a)
    a, b = b, a + b
print(*res)'

# Liam: primes (wrong)
submit "$TOK_S5" "$EX_PRIMES" "python" 'n = int(input())
for i in range(2, n+1):
    if all(i % j != 0 for j in range(2, i)):
        print(i)'

# Liam: hello world
submit "$TOK_S5" "$EX_HELLO" "python" 'print("Hello, World!")'

# Maya: merge sort
submit "$TOK_S6" "$EX_SORT" "python" 'n = int(input())
arr = list(map(int, input().split()))
print(*sorted(arr))'

# Noah: JS reverse
submit "$TOK_S7" "$EX_JS_REVERSE" "javascript" 'const rl = require("readline").createInterface({input:process.stdin});
rl.on("line", l => { console.log(l.split("").reverse().join("")); rl.close(); });'

# Olivia: Go hello
submit "$TOK_S8" "$EX_GO_HELLO" "go" 'package main
import "fmt"
func main() { fmt.Println("Hello, Gopher!") }'

# Olivia: valid parentheses
submit "$TOK_S8" "$EX_STACK" "python" 's = input()
stack = []
m = {")":"(","}":"{","]":"["}
for c in s:
    if c in "({[": stack.append(c)
    elif c in m:
        if not stack or stack[-1] != m[c]:
            print("INVALID"); exit()
        stack.pop()
print("VALID" if not stack else "INVALID")'

log ""
log "══════════════════════════════════════════════════"
log " Seed complete!"
log "══════════════════════════════════════════════════"
log ""
log " Accounts (all passwords: Password123)"
log "  admin@cs.edu                  — admin"
log "  alice.morgan@cs.edu           — professor  (Python, Algo courses)"
log "  bob.carter@cs.edu             — professor  (Go, C++ courses)"
log "  carol.james@cs.edu            — professor  (Web course)"
log "  david.kim@cs.edu              — TA         (Python, Algo)"
log "  emma.silva@cs.edu             — TA         (Python, Web)"
log "  frank.liu@cs.edu              — TA         (Algo)"
log "  grace.patel@cs.edu            — TA         (Go)"
log "  henry.walsh@student.edu       — student"
log "  isabella.chen@student.edu     — student"
log "  jack.novak@student.edu        — student"
log "  kate.torres@student.edu       — student"
log "  liam.oconnor@student.edu      — student"
log "  maya.singh@student.edu        — student"
log "  noah.berg@student.edu         — student"
log "  olivia.martin@student.edu     — student"
log ""
log " Courses"
log "  • Introduction to Python Programming  (published, 6 exercises)"
log "  • Data Structures & Algorithms        (published, 4 exercises)"
log "  • Systems Programming in Go           (published, 3 exercises)"
log "  • Web Development Fundamentals        (published, 3 exercises)"
log "  • C++ for Competitive Programming     (unpublished draft)"
