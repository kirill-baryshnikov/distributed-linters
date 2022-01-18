#!/usr/bin/env python3

from termcolor import colored
from contextlib import contextmanager
from time import sleep
from urllib import request
import json
import psutil
import subprocess

manager_url = "http://localhost:10000"


def info(text):
    print(colored(text, "magenta"))


def test(text):
    print()
    print(colored(text, "cyan", attrs=["bold"]))


def ok():
    print(colored("OK", "green", attrs=["bold"]))


@contextmanager
def process(*args, **kwargs):
    proc = subprocess.Popen(*args, **kwargs)
    try:
        yield proc
    finally:
        for child in psutil.Process(proc.pid).children(recursive=True):
            child.kill()
        proc.kill()


def test_python_linter():
    code_to_lint = "int main(){}"
    endpoint = "{}/v1/lint/python".format(manager_url)

    req = request.Request(endpoint, method="POST")
    req.add_header('Content-Type', 'application/json')
    data = json.dumps({"content": code_to_lint}).encode()
    resp = request.urlopen(req, data=data)

    content = json.loads(resp.read())
    if not content["result"]:
        raise Exception("Invalid linter response")


def test_java_linter():
    code_to_lint = "int main(){}"
    endpoint = "{}/v1/lint/java".format(manager_url)

    req = request.Request(endpoint, method="POST")
    req.add_header('Content-Type', 'application/json')
    data = json.dumps({"content": code_to_lint}).encode()
    resp = request.urlopen(req, data=data)

    content = json.loads(resp.read())
    if not content["result"]:
        raise Exception("Invalid linter response")


def remove_python_linter():
    endpoint = "{}/v1/admin/workers/python".format(manager_url)

    req = request.Request(endpoint, method="DELETE")
    request.urlopen(req)


def add_python_linter():
    endpoint = "{}/v1/admin/workers/python".format(manager_url)

    req = request.Request(endpoint, method="POST")
    request.urlopen(req)


def new_python_linter_version():
    new_version = "bin/python-linter-2.0"
    endpoint = "{}/v1/admin/version/python".format(manager_url)

    req = request.Request(endpoint, method="POST")
    req.add_header('Content-Type', 'application/json')
    data = json.dumps({"version": new_version}).encode()
    request.urlopen(req, data=data)


def count_python_linters(manager):
    count = 0
    for child in psutil.Process(manager.pid).children(recursive=True):
        if "python-linter" in child.name():
            count += 1
    return count


def count_python_linters_with_new_version(manager):
    count = 0
    for child in psutil.Process(manager.pid).children(recursive=True):
        if "python-linter-2.0" in child.name():
            count += 1
    return count


def print_active_linters(manager):
    s = colored("Active linters:\n", "magenta")
    for child in psutil.Process(manager.pid).children(recursive=True):
        s += "    " + colored(child.name(), "magenta") + "\n"
    print(s, end="")


def kill_python_linters(manager):
    for child in psutil.Process(manager.pid).children(recursive=True):
        if "python-linter" in child.name():
            child.kill()


def kill_java_linter(manager):
    for child in psutil.Process(manager.pid).children(recursive=True):
        if "java-linter" in child.name():
            child.kill()
            break


info("Starting manager")
with process(["bin/manager"]) as manager:
    info("Waiting for manager to initialize")
    sleep(10)
    info("Starting tests")

    test("Testing python linter")
    test_python_linter()
    ok()

    test("Testing crash resistance")
    print_active_linters(manager)
    info("Killing one java linter")
    kill_java_linter(manager)
    print_active_linters(manager)
    info("Wait for manager to react")
    sleep(1)

    info("Testing java linter")
    test_java_linter()
    ok()

    test("Testing restarting linters")
    info("Killing all python linters")
    kill_python_linters(manager)
    print_active_linters(manager)

    info("Wait for manager to react")
    sleep(10)
    print_active_linters(manager)

    info("Testing python linter")
    test_python_linter()
    ok()

    test("Testing python worker removal")
    c1 = count_python_linters(manager)
    info("Python worker count before removal: {}".format(c1))
    remove_python_linter()
    info("Waiting for manager to react")

    sleep(10)
    c2 = count_python_linters(manager)
    info("Python worker count after removal: {}".format(c2))
    if not c1 == c2 + 1:
        raise Exception("Worker count did not decrease")
    ok()

    test("Testing python worker addition")
    c1 = count_python_linters(manager)
    info("Python worker count before addition: {}".format(c1))
    add_python_linter()
    info("Waiting for manager to react")

    sleep(10)
    c2 = count_python_linters(manager)
    info("Python worker count after addition: {}".format(c2))
    if not c1 + 1 == c2:
        raise Exception("Worker count did not increase")
    ok()

    test("Testing python worker update")
    new_python_linter_version()
    for i in range(5):
        sleep(5)
        print_active_linters(manager)
    if not c2 == count_python_linters_with_new_version(manager):
        raise Exception("Uddate failed")
    ok()

print(colored("All tests have passed", "green", attrs=["bold"]))
