#!/usr/bin/env python3
"""Seed the CS Educational Platform database with realistic test data."""

import json
import os
import subprocess
import sys
import tempfile
import textwrap
from pathlib import Path
from typing import Optional

try:
    import urllib.request
    import urllib.error
except ImportError:
    sys.exit("Python 3 stdlib required")

API = "http://localhost:8080"

if sys.platform == "win32":
    try:
        import ctypes
        ctypes.windll.kernel32.SetConsoleMode(ctypes.windll.kernel32.GetStdHandle(-11), 7)
        GREEN = "\033[0;32m"
        YELLOW = "\033[1;33m"
        RED = "\033[0;31m"
        NC = "\033[0m"
    except Exception:
        GREEN = YELLOW = RED = NC = ""
else:
    GREEN = "\033[0;32m"
    YELLOW = "\033[1;33m"
    RED = "\033[0;31m"
    NC = "\033[0m"


def log(msg: str) -> None:
    print(f"{GREEN}[seed]{NC} {msg}")


def warn(msg: str) -> None:
    print(f"{YELLOW}[warn]{NC} {msg}", file=sys.stderr)


def api(method: str, path: str, body: dict | None = None, token: str | None = None) -> dict | None:
    url = f"{API}{path}"
    data = json.dumps(body).encode() if body else None
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req) as resp:
            text = resp.read()
            return json.loads(text) if text else None
    except urllib.error.HTTPError as e:
        msg = e.read().decode()
        warn(f"{method} {path} → {e.code}: {msg}")
        return None


def register(email: str, password: str, first: str, last: str, role: str) -> tuple[str, str]:
    """Returns (token, user_id)."""
    resp = api("POST", "/auth/register", {
        "email": email, "password": password,
        "first_name": first, "last_name": last, "role": role,
    })
    if resp:
        return resp["token"], resp["user_id"]
    # Already registered — log in instead
    resp = api("POST", "/auth/login", {"email": email, "password": password})
    if resp:
        return resp["token"], resp["user_id"]
    sys.exit(f"Could not register or log in {email}")


def create_course(token: str, title: str, description: str, published: bool) -> Optional[str]:
    resp = api("POST", "/courses", {
        "title": title, "description": description, "is_published": published,
    }, token)
    return resp["id"] if resp else None


def create_exercise(
    token: str, course_id: str, title: str, description: str,
    instructions: str, difficulty: str, language: str, template: str,
    time_ms: int, mem_kb: int, published: bool,
) -> Optional[str]:
    resp = api("POST", f"/courses/{course_id}/exercises", {
        "title": title,
        "description": description,
        "instructions": instructions,
        "difficulty": difficulty,
        "language": language,
        "template_code": template,
        "time_limit_ms": time_ms,
        "memory_limit_kb": mem_kb,
        "is_published": published,
    }, token)
    return resp["id"] if resp else None


def add_test_case(token: str, exercise_id: str, inp: str, expected: str, hidden: bool, ordinal: int) -> None:
    api("POST", f"/exercises/{exercise_id}/test-cases", {
        "input": inp, "expected_output": expected,
        "is_hidden": hidden, "ordinal": ordinal,
    }, token)


def enroll(token: str, course_id: str) -> None:
    api("POST", f"/courses/{course_id}/enroll", {}, token)


def promote_to_ta(prof_token: str, course_id: str, user_id: str) -> None:
    """Promote an enrolled student to course TA. Must be called with the professor's token."""
    api("PATCH", f"/courses/{course_id}/members/{user_id}/role",
        {"role": "teaching_assistant"}, prof_token)


def submit(token: str, exercise_id: str, language: str, code: str) -> None:
    api("POST", f"/exercises/{exercise_id}/submit", {
        "language": language, "code": code,
    }, token)


# ── 1. Create users ──────────────────────────────────────────────────────────

log("Creating professors…")
T_PROF1, _ = register("alice.morgan@cs.edu",   "Password123", "Alice",    "Morgan",   "professor")
T_PROF2, _ = register("bob.carter@cs.edu",     "Password123", "Bob",      "Carter",   "professor")
T_PROF3, _ = register("carol.james@cs.edu",    "Password123", "Carol",    "James",    "professor")

# TAs are regular students globally; they get promoted to course TA per-course below.
log("Creating future TAs (registered as student)…")
T_TA1, UID_TA1 = register("david.kim@cs.edu",   "Password123", "David",  "Kim",   "student")
T_TA2, UID_TA2 = register("emma.silva@cs.edu",  "Password123", "Emma",   "Silva", "student")
T_TA3, UID_TA3 = register("frank.liu@cs.edu",   "Password123", "Frank",  "Liu",   "student")
T_TA4, UID_TA4 = register("grace.patel@cs.edu", "Password123", "Grace",  "Patel", "student")

log("Creating students…")
T_S1, _ = register("henry.walsh@student.edu",     "Password123", "Henry",   "Walsh",   "student")
T_S2, _ = register("isabella.chen@student.edu",   "Password123", "Isabella","Chen",    "student")
T_S3, _ = register("jack.novak@student.edu",      "Password123", "Jack",    "Novak",   "student")
T_S4, _ = register("kate.torres@student.edu",     "Password123", "Kate",    "Torres",  "student")
T_S5, _ = register("liam.oconnor@student.edu",    "Password123", "Liam",    "OConnor", "student")
T_S6, _ = register("maya.singh@student.edu",      "Password123", "Maya",    "Singh",   "student")
T_S7, _ = register("noah.berg@student.edu",       "Password123", "Noah",    "Berg",    "student")
T_S8, _ = register("olivia.martin@student.edu",   "Password123", "Olivia",  "Martin",  "student")

log("Creating admin user via direct DB insert…")
# Generate bcrypt hash using the project's Go toolchain (cost 12)
BACKEND_DIR = str(Path(__file__).resolve().parent.parent / "backend")
GENHASH = os.path.join(tempfile.gettempdir(), "genhash_seed.go")
with open(GENHASH, "w") as f:
    f.write('package main\nimport ("fmt";"golang.org/x/crypto/bcrypt")\n'
            'func main(){h,_:=bcrypt.GenerateFromPassword([]byte("Password123"),12);fmt.Println(string(h))}\n')
gen = subprocess.run(["go", "run", GENHASH], capture_output=True, text=True, cwd=BACKEND_DIR)
ADMIN_HASH = gen.stdout.strip()
if not ADMIN_HASH:
    sys.exit(f"Failed to generate bcrypt hash: {gen.stderr}")
result = subprocess.run(
    [
        "docker", "exec", "csplatform_db",
        "psql", "-U", "csplatform", "-d", "csplatform_db", "-c",
        f"INSERT INTO users (email, password_hash, first_name, last_name, role) "
        f"VALUES ('admin@cs.edu', '{ADMIN_HASH}', 'Admin', 'User', 'admin') "
        f"ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;",
    ],
    capture_output=True, text=True,
)
if result.returncode != 0:
    warn(f"Admin insert: {result.stderr.strip()}")
else:
    log("Admin: admin@cs.edu / Password123")

log("All users ready.")

# ── 2. Create courses ────────────────────────────────────────────────────────

log("Creating courses…")

CID_PYTHON = create_course(T_PROF1,
    "Introduction to Python Programming",
    "A beginner-friendly course covering Python fundamentals: variables, loops, "
    "functions, and basic data structures. No prior experience required.",
    True,
)

CID_ALGO = create_course(T_PROF1,
    "Data Structures & Algorithms",
    "Deep dive into arrays, linked lists, trees, graphs, sorting, and searching. "
    "Prepares students for technical interviews and competitive programming.",
    True,
)

CID_GO = create_course(T_PROF2,
    "Systems Programming in Go",
    "Concurrent and systems-level programming using Go. Topics: goroutines, channels, "
    "interfaces, error handling, and the standard library.",
    True,
)

CID_WEB = create_course(T_PROF3,
    "Web Development Fundamentals",
    "Full-stack basics with HTML, CSS, and JavaScript. Students build and deploy "
    "small interactive web applications.",
    True,
)

CID_CPP = create_course(T_PROF2,
    "C++ for Competitive Programming",
    "DRAFT — STL, templates, and advanced problem-solving techniques. "
    "Not yet published.",
    False,
)

log("Courses created.")

# ── 3. Enroll students and TAs ───────────────────────────────────────────────

log("Enrolling members…")

# Python — largest class
for tok in [T_S1, T_S2, T_S3, T_S4, T_S5, T_S6, T_S7, T_S8, T_TA1, T_TA2]:
    enroll(tok, CID_PYTHON)
promote_to_ta(T_PROF1, CID_PYTHON, UID_TA1)
promote_to_ta(T_PROF1, CID_PYTHON, UID_TA2)

# Algorithms
for tok in [T_S1, T_S2, T_S3, T_S5, T_S7, T_TA1, T_TA3]:
    enroll(tok, CID_ALGO)
promote_to_ta(T_PROF1, CID_ALGO, UID_TA1)
promote_to_ta(T_PROF1, CID_ALGO, UID_TA3)

# Go
for tok in [T_S2, T_S4, T_S6, T_S8, T_TA4]:
    enroll(tok, CID_GO)
promote_to_ta(T_PROF2, CID_GO, UID_TA4)

# Web
for tok in [T_S3, T_S5, T_S6, T_S7, T_S8, T_TA2, T_TA3]:
    enroll(tok, CID_WEB)
promote_to_ta(T_PROF3, CID_WEB, UID_TA2)
promote_to_ta(T_PROF3, CID_WEB, UID_TA3)

log("Enrollment done.")

# ── 4. Python exercises ──────────────────────────────────────────────────────

log("Creating Python exercises…")

EX_HELLO = create_exercise(T_PROF1, CID_PYTHON,
    "Hello, World!",
    "Your very first Python program.",
    'Write a Python program that prints exactly:\nHello, World!',
    "easy", "python",
    'print("Hello, World!")',
    2000, 65536, True,
)
add_test_case(T_PROF1, EX_HELLO, "",  "Hello, World!", False, 1)
add_test_case(T_PROF1, EX_HELLO, "",  "Hello, World!", True,  2)

EX_SUM = create_exercise(T_PROF1, CID_PYTHON,
    "Sum of Two Numbers",
    "Read two integers from stdin and print their sum.",
    "Read two integers from stdin (one per line) and print their sum on a single line.",
    "easy", "python",
    "a = int(input())\nb = int(input())\nprint(a + b)",
    2000, 65536, True,
)
add_test_case(T_PROF1, EX_SUM, "3\n5",         "8",       False, 1)
add_test_case(T_PROF1, EX_SUM, "0\n0",         "0",       False, 2)
add_test_case(T_PROF1, EX_SUM, "-7\n4",        "-3",      True,  3)
add_test_case(T_PROF1, EX_SUM, "1000000\n999999", "1999999", True, 4)

EX_FIZ = create_exercise(T_PROF1, CID_PYTHON,
    "FizzBuzz",
    "The classic FizzBuzz problem.",
    'Read an integer N from stdin. Print numbers from 1 to N (inclusive), one per line. '
    'Replace multiples of 3 with "Fizz", multiples of 5 with "Buzz", multiples of both with "FizzBuzz".',
    "easy", "python",
    textwrap.dedent("""\
        n = int(input())
        for i in range(1, n + 1):
            if i % 15 == 0:
                print("FizzBuzz")
            elif i % 3 == 0:
                print("Fizz")
            elif i % 5 == 0:
                print("Buzz")
            else:
                print(i)"""),
    2000, 65536, True,
)
add_test_case(T_PROF1, EX_FIZ, "15",
    "1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz",
    False, 1)
add_test_case(T_PROF1, EX_FIZ, "5",  "1\n2\nFizz\n4\nBuzz", False, 2)
add_test_case(T_PROF1, EX_FIZ, "1",  "1",                    True,  3)

EX_PALINDROME = create_exercise(T_PROF1, CID_PYTHON,
    "Palindrome Check",
    "Determine whether a string is a palindrome.",
    'Read a single line of text. Print "YES" if it is a palindrome (case-insensitive, '
    'ignore spaces), "NO" otherwise.',
    "medium", "python",
    textwrap.dedent("""\
        s = input().replace(" ", "").lower()
        print("YES" if s == s[::-1] else "NO")"""),
    2000, 65536, True,
)
add_test_case(T_PROF1, EX_PALINDROME, "racecar",                   "YES", False, 1)
add_test_case(T_PROF1, EX_PALINDROME, "hello",                     "NO",  False, 2)
add_test_case(T_PROF1, EX_PALINDROME, "A man a plan a canal Panama","YES", True,  3)
add_test_case(T_PROF1, EX_PALINDROME, "Was it a car or a cat I saw","YES", True,  4)

EX_PRIMES = create_exercise(T_PROF1, CID_PYTHON,
    "Sieve of Eratosthenes",
    "Find all prime numbers up to N.",
    "Read an integer N from stdin. Print all prime numbers up to N (inclusive), "
    "one per line, in ascending order.",
    "medium", "python",
    textwrap.dedent("""\
        n = int(input())
        sieve = [True] * (n + 1)
        sieve[0] = sieve[1] = False
        for i in range(2, int(n**0.5) + 1):
            if sieve[i]:
                for j in range(i*i, n+1, i):
                    sieve[j] = False
        for i in range(2, n+1):
            if sieve[i]:
                print(i)"""),
    3000, 65536, True,
)
add_test_case(T_PROF1, EX_PRIMES, "10", "2\n3\n5\n7",                          False, 1)
add_test_case(T_PROF1, EX_PRIMES, "2",  "2",                                    False, 2)
add_test_case(T_PROF1, EX_PRIMES, "3",  "2\n3",                                 True,  3)
add_test_case(T_PROF1, EX_PRIMES, "50", "2\n3\n5\n7\n11\n13\n17\n19\n23\n29\n31\n37\n41\n43\n47", True, 4)

EX_FIBO = create_exercise(T_PROF1, CID_PYTHON,
    "Fibonacci Sequence",
    "Generate the first N Fibonacci numbers.",
    "Read an integer N from stdin. Print the first N Fibonacci numbers separated by "
    "spaces on one line. F(1)=1, F(2)=1.",
    "easy", "python",
    textwrap.dedent("""\
        n = int(input())
        a, b = 1, 1
        result = []
        for _ in range(n):
            result.append(str(a))
            a, b = b, a + b
        print(" ".join(result))"""),
    2000, 65536, True,
)
add_test_case(T_PROF1, EX_FIBO, "1",  "1",                  False, 1)
add_test_case(T_PROF1, EX_FIBO, "5",  "1 1 2 3 5",          False, 2)
add_test_case(T_PROF1, EX_FIBO, "10", "1 1 2 3 5 8 13 21 34 55", True, 3)

# ── 5. Algorithms exercises ──────────────────────────────────────────────────

log("Creating Algorithms exercises…")

EX_BSEARCH = create_exercise(T_PROF1, CID_ALGO,
    "Binary Search",
    "Implement binary search on a sorted array.",
    "First line: N (size of array).\nSecond line: N space-separated integers in ascending order.\n"
    "Third line: target integer.\nPrint the 0-based index where target is found, or -1 if not present.",
    "easy", "python",
    textwrap.dedent("""\
        n = int(input())
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
        print(result)"""),
    2000, 65536, True,
)
add_test_case(T_PROF1, EX_BSEARCH, "5\n1 3 5 7 9\n7", "3",  False, 1)
add_test_case(T_PROF1, EX_BSEARCH, "5\n1 3 5 7 9\n4", "-1", False, 2)
add_test_case(T_PROF1, EX_BSEARCH, "1\n42\n42",        "0",  True,  3)

EX_SORT = create_exercise(T_PROF1, CID_ALGO,
    "Merge Sort",
    "Implement merge sort and sort an array of integers.",
    "First line: N.\nSecond line: N space-separated integers.\nPrint the sorted array on one line, space-separated.",
    "medium", "python",
    textwrap.dedent("""\
        def merge_sort(arr):
            if len(arr) <= 1:
                return arr
            mid = len(arr) // 2
            left = merge_sort(arr[:mid])
            right = merge_sort(arr[mid:])
            result, i, j = [], 0, 0
            while i < len(left) and j < len(right):
                if left[i] <= right[j]:
                    result.append(left[i]); i += 1
                else:
                    result.append(right[j]); j += 1
            return result + left[i:] + right[j:]

        n = int(input())
        arr = list(map(int, input().split()))
        print(*merge_sort(arr))"""),
    3000, 65536, True,
)
add_test_case(T_PROF1, EX_SORT, "5\n3 1 4 1 5",   "1 1 3 4 5",     False, 1)
add_test_case(T_PROF1, EX_SORT, "1\n7",            "7",             False, 2)
add_test_case(T_PROF1, EX_SORT, "6\n-3 0 5 -1 2 8","-3 -1 0 2 5 8",True,  3)

EX_STACK = create_exercise(T_PROF1, CID_ALGO,
    "Valid Parentheses",
    "Check if a string of brackets is valid using a stack.",
    'Read a string containing only "(", ")", "{", "}", "[", "]".\n'
    'Print "VALID" if brackets are properly nested and closed, "INVALID" otherwise.',
    "medium", "python",
    textwrap.dedent("""\
        s = input()
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
        print("VALID" if not stack else "INVALID")"""),
    2000, 65536, True,
)
add_test_case(T_PROF1, EX_STACK, "(){}[]", "VALID",   False, 1)
add_test_case(T_PROF1, EX_STACK, "([)]",   "INVALID", False, 2)
add_test_case(T_PROF1, EX_STACK, "{[]}",   "VALID",   True,  3)
add_test_case(T_PROF1, EX_STACK, "(((",    "INVALID", True,  4)

EX_BFS = create_exercise(T_PROF1, CID_ALGO,
    "BFS Shortest Path",
    "Find the shortest path in an unweighted undirected graph using BFS.",
    "First line: N M (nodes 1..N, M edges).\n"
    "Next M lines: u v (undirected edge).\n"
    "Last line: S T (source and target).\n"
    "Print the shortest path length, or -1 if unreachable.",
    "hard", "python",
    textwrap.dedent("""\
        from collections import deque
        n, m = map(int, input().split())
        adj = [[] for _ in range(n + 1)]
        for _ in range(m):
            u, v = map(int, input().split())
            adj[u].append(v); adj[v].append(u)
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
        print(dist[t])"""),
    3000, 131072, True,
)
add_test_case(T_PROF1, EX_BFS, "4 4\n1 2\n2 3\n3 4\n1 3\n1 4", "2",  False, 1)
add_test_case(T_PROF1, EX_BFS, "3 1\n1 2\n1 3",                 "-1", False, 2)
add_test_case(T_PROF1, EX_BFS, "2 1\n1 2\n1 2",                 "1",  True,  3)

# ── 6. Go exercises ──────────────────────────────────────────────────────────

log("Creating Go exercises…")

EX_GO_HELLO = create_exercise(T_PROF2, CID_GO,
    "Hello, Gopher!",
    "Your first Go program.",
    'Write a Go program that prints exactly:\nHello, Gopher!',
    "easy", "go",
    textwrap.dedent("""\
        package main

        import "fmt"

        func main() {
            fmt.Println("Hello, Gopher!")
        }"""),
    3000, 65536, True,
)
add_test_case(T_PROF2, EX_GO_HELLO, "", "Hello, Gopher!", False, 1)
add_test_case(T_PROF2, EX_GO_HELLO, "", "Hello, Gopher!", True,  2)

EX_GO_SUM = create_exercise(T_PROF2, CID_GO,
    "Sum with Goroutines",
    "Split an array sum across two goroutines and combine results via a channel.",
    "Read N from stdin, then N integers. Split the array in half, sum each half in a separate "
    "goroutine, send results on a channel, and print the total sum.",
    "medium", "go",
    textwrap.dedent("""\
        package main

        import "fmt"

        func sumHalf(nums []int, ch chan int) {
            total := 0
            for _, v := range nums {
                total += v
            }
            ch <- total
        }

        func main() {
            var n int
            fmt.Scan(&n)
            nums := make([]int, n)
            for i := range nums {
                fmt.Scan(&nums[i])
            }
            ch := make(chan int, 2)
            go sumHalf(nums[:n/2], ch)
            go sumHalf(nums[n/2:], ch)
            fmt.Println(<-ch + <-ch)
        }"""),
    3000, 65536, True,
)
add_test_case(T_PROF2, EX_GO_SUM, "5\n1 2 3 4 5", "15",  False, 1)
add_test_case(T_PROF2, EX_GO_SUM, "4\n-1 0 1 2",  "2",   False, 2)
add_test_case(T_PROF2, EX_GO_SUM, "1\n42",         "42",  True,  3)

EX_GO_CHAN = create_exercise(T_PROF2, CID_GO,
    "Fan-out with Channels",
    "Distribute work across goroutines using a jobs/results channel pattern.",
    "Read N from stdin. Send integers 1..N to a jobs channel. Have 3 worker goroutines "
    "double each value and send results to a results channel. "
    "Print all results sorted ascending on one line, space-separated.",
    "hard", "go",
    textwrap.dedent("""\
        package main

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
                if i > 0 {
                    fmt.Print(" ")
                }
                fmt.Print(v)
            }
            fmt.Println()
        }"""),
    5000, 131072, True,
)
add_test_case(T_PROF2, EX_GO_CHAN, "5", "2 4 6 8 10", False, 1)
add_test_case(T_PROF2, EX_GO_CHAN, "1", "2",          True,  2)

# ── 7. JavaScript exercises ──────────────────────────────────────────────────

log("Creating JavaScript exercises…")

EX_JS_HELLO = create_exercise(T_PROF3, CID_WEB,
    "Hello from Node",
    "Print a greeting from a Node.js script.",
    'Write a JavaScript program that prints:\nHello from Node.js!',
    "easy", "javascript",
    'console.log("Hello from Node.js!");',
    2000, 65536, True,
)
add_test_case(T_PROF3, EX_JS_HELLO, "", "Hello from Node.js!", False, 1)

EX_JS_REVERSE = create_exercise(T_PROF3, CID_WEB,
    "Reverse a String",
    "Reverse a string using JavaScript.",
    "Read a line from stdin and print it reversed.",
    "easy", "javascript",
    textwrap.dedent("""\
        const readline = require("readline");
        const rl = readline.createInterface({ input: process.stdin });
        rl.on("line", (line) => {
            console.log(line.split("").reverse().join(""));
            rl.close();
        });"""),
    2000, 65536, True,
)
add_test_case(T_PROF3, EX_JS_REVERSE, "hello",   "olleh",   False, 1)
add_test_case(T_PROF3, EX_JS_REVERSE, "OpenAI",  "IAnepO",  False, 2)
add_test_case(T_PROF3, EX_JS_REVERSE, "racecar", "racecar", True,  3)

EX_JS_REDUCE = create_exercise(T_PROF3, CID_WEB,
    "Array Sum with reduce()",
    "Use Array.prototype.reduce to sum an array.",
    "Read N, then N space-separated numbers. Use reduce() to compute and print their sum.",
    "medium", "javascript",
    textwrap.dedent("""\
        const lines = [];
        require("readline").createInterface({ input: process.stdin })
          .on("line", l => lines.push(l.trim()))
          .on("close", () => {
              const nums = lines[1].split(" ").map(Number);
              console.log(nums.reduce((acc, x) => acc + x, 0));
          });"""),
    2000, 65536, True,
)
add_test_case(T_PROF3, EX_JS_REDUCE, "4\n1 2 3 4", "10", False, 1)
add_test_case(T_PROF3, EX_JS_REDUCE, "3\n-1 0 1",  "0",  True,  2)

EX_JS_PROMISE = create_exercise(T_PROF3, CID_WEB,
    "Promise Chain",
    "Chain multiple .then() handlers to transform a value.",
    "No stdin needed. Start with value 1, double it, add 10, and multiply by 3, "
    "all through .then() chains. Print the final result.",
    "medium", "javascript",
    textwrap.dedent("""\
        Promise.resolve(1)
          .then(v => v * 2)
          .then(v => v + 10)
          .then(v => v * 3)
          .then(v => console.log(v));"""),
    2000, 65536, True,
)
add_test_case(T_PROF3, EX_JS_PROMISE, "", "36", False, 1)
add_test_case(T_PROF3, EX_JS_PROMISE, "", "36", True,  2)

# ── 8. Sample submissions ────────────────────────────────────────────────────

log("Creating sample submissions…")

# Henry: correct FizzBuzz
submit(T_S1, EX_FIZ, "python", textwrap.dedent("""\
    n = int(input())
    for i in range(1, n + 1):
        if i % 15 == 0: print("FizzBuzz")
        elif i % 3 == 0: print("Fizz")
        elif i % 5 == 0: print("Buzz")
        else: print(i)"""))

# Henry: palindrome (ignores spaces — wrong answer)
submit(T_S1, EX_PALINDROME, "python", 's = input().lower()\nprint("YES" if s == s[::-1] else "NO")')

# Henry: hello world
submit(T_S1, EX_HELLO, "python", 'print("Hello, World!")')

# Isabella: correct sum
submit(T_S2, EX_SUM, "python", "a = int(input())\nb = int(input())\nprint(a + b)")

# Isabella: binary search
submit(T_S2, EX_BSEARCH, "python", textwrap.dedent("""\
    n = int(input())
    arr = list(map(int, input().split()))
    target = int(input())
    print(arr.index(target) if target in arr else -1)"""))

# Jack: correct Python hello (Jack enrolled in Python and Algo, not Go)
submit(T_S3, EX_HELLO, "python", 'print("Hello, World!")')

# Jack: valid parentheses
submit(T_S3, EX_STACK, "python", textwrap.dedent("""\
    s = input()
    stack = []
    m = {")":"(","}":"{","]":"["}
    for c in s:
        if c in "({[": stack.append(c)
        elif c in m:
            if not stack or stack[-1] != m[c]:
                print("INVALID"); exit()
            stack.pop()
    print("VALID" if not stack else "INVALID")"""))

# Kate: Fibonacci
submit(T_S4, EX_FIBO, "python", textwrap.dedent("""\
    n = int(input())
    a, b = 1, 1
    res = []
    for _ in range(n):
        res.append(a)
        a, b = b, a + b
    print(*res)"""))

# Liam: primes (brute force — slow but correct for small N)
submit(T_S5, EX_PRIMES, "python", textwrap.dedent("""\
    n = int(input())
    for i in range(2, n+1):
        if all(i % j != 0 for j in range(2, i)):
            print(i)"""))

# Liam: hello world
submit(T_S5, EX_HELLO, "python", 'print("Hello, World!")')

# Maya: JS reverse (Maya enrolled in Web)
submit(T_S6, EX_JS_REVERSE, "javascript", textwrap.dedent("""\
    const rl = require("readline").createInterface({input:process.stdin});
    rl.on("line", l => { console.log(l.split("").reverse().join("")); rl.close(); });"""))

# Maya: JS promise (Maya enrolled in Web)
submit(T_S6, EX_JS_PROMISE, "javascript", textwrap.dedent("""\
    Promise.resolve(1)
      .then(v => v * 2)
      .then(v => v + 10)
      .then(v => v * 3)
      .then(v => console.log(v));"""))

# Noah: BFS
submit(T_S7, EX_BFS, "python", textwrap.dedent("""\
    from collections import deque
    n, m = map(int, input().split())
    adj = [[] for _ in range(n + 1)]
    for _ in range(m):
        u, v = map(int, input().split())
        adj[u].append(v); adj[v].append(u)
    s, t = map(int, input().split())
    dist = [-1] * (n + 1); dist[s] = 0
    q = deque([s])
    while q:
        u = q.popleft()
        for v in adj[u]:
            if dist[v] == -1:
                dist[v] = dist[u] + 1; q.append(v)
    print(dist[t])"""))

# Olivia: correct Go hello
submit(T_S8, EX_GO_HELLO, "go", 'package main\nimport "fmt"\nfunc main() { fmt.Println("Hello, Gopher!") }')

# Olivia: Go goroutine sum
submit(T_S8, EX_GO_SUM, "go", textwrap.dedent("""\
    package main
    import "fmt"
    func sumHalf(nums []int, ch chan int) {
        total := 0
        for _, v := range nums { total += v }
        ch <- total
    }
    func main() {
        var n int; fmt.Scan(&n)
        nums := make([]int, n)
        for i := range nums { fmt.Scan(&nums[i]) }
        ch := make(chan int, 2)
        go sumHalf(nums[:n/2], ch)
        go sumHalf(nums[n/2:], ch)
        fmt.Println(<-ch + <-ch)
    }"""))

# ── Done ─────────────────────────────────────────────────────────────────────

print()
log("=" * 54)
log(" Seed complete!")
log("=" * 54)
print()
log("Accounts (all passwords: Password123)")
log("  admin@cs.edu                   — admin")
log("  alice.morgan@cs.edu            — professor (Python, Algo)")
log("  bob.carter@cs.edu              — professor (Go, C++)")
log("  carol.james@cs.edu             — professor (Web)")
log("  david.kim@cs.edu               — student / course TA in: Python, Algo")
log("  emma.silva@cs.edu              — student / course TA in: Python, Web")
log("  frank.liu@cs.edu               — student / course TA in: Algo, Web")
log("  grace.patel@cs.edu             — student / course TA in: Go")
log("  henry.walsh@student.edu        — student")
log("  isabella.chen@student.edu      — student")
log("  jack.novak@student.edu         — student")
log("  kate.torres@student.edu        — student")
log("  liam.oconnor@student.edu       — student")
log("  maya.singh@student.edu         — student")
log("  noah.berg@student.edu          — student")
log("  olivia.martin@student.edu      — student")
print()
log("Courses")
log("  Introduction to Python Programming  (published, 6 exercises)")
log("  Data Structures & Algorithms        (published, 4 exercises)")
log("  Systems Programming in Go           (published, 3 exercises)")
log("  Web Development Fundamentals        (published, 4 exercises)")
log("  C++ for Competitive Programming     (unpublished draft)")
