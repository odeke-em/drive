#!/usr/bin/env python
# Copyright 2015 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import hashlib
import os
import re
import subprocess
import sys
import traceback
from contextlib import contextmanager

# config
g_testdir = 'testdir'
g_drive_bin = 'drive'

g_fail_stop = False

# global variables
count_ok = 0
count_bad = 0

g_driveignore_fn = os.path.join(g_testdir, '.driveignore')


def init():
    if not os.path.exists(g_drive_bin):
        print 'Drive executable (path=%s) not found' % repr(g_drive_bin)
        sys.exit(1)
    if not os.path.exists(os.path.join(g_testdir, '.gd')):
        print 'Please init drive folder %s/ first' % g_testdir
        sys.exit(1)
    if Drive.list('') != []:
        print 'Warning! This test tool has destructive side-effect and will erase you drive'
        sys.exit(1)


def expect_eq(expected, actual):
    global count_ok, count_bad
    if expected != actual:
        print '[expected]', repr(expected)
        print '[actual]', repr(actual)
        print 'failed'
        count_bad += 1
        raise Exception
    else:
        count_ok += 1


def expect_ne(not_expected, actual):
    global count_ok, count_bad
    if not_expected == actual:
        print '[not expected equal]', repr(actual)
        print 'failed'
        count_bad += 1
        raise Exception
    else:
        count_ok += 1


def expect_true(actual):
    expect_eq(True, bool(actual))


class Drive:
    @classmethod
    def run(cls, cmd, *args, **argd):
        extraflags = argd.get('extraflags', [])
        if type(extraflags) not in (list, tuple):
            extraflags = [extraflags]
        cmd = [g_drive_bin] + [cmd] + list(extraflags) + list(args)

        #print '$',
        if argd.get('input') is not None:
            if re.match(r'^[\x32-\x79\n]+$', argd.get('input')):
                print 'echo "%s" |' % argd.get('input'),
            else:
                print 'echo ... |',
        print subprocess.list2cmdline(cmd)

        try:
            cwd = os.getcwd()
            os.chdir(g_testdir)
            if argd.get('input') is None:
                p = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            else:
                p = subprocess.Popen(cmd, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        finally:
            os.chdir(cwd)

        out, err = p.communicate(argd.get('input'))

        return p.returncode, out, err

    @classmethod
    def run_ok(cls, *args, **argd):
        returncode, out, err = cls.run(*args, **argd)
        if returncode != 0:
            if out != '':
                print '[stdout]'
                sys.stdout.write(out)
            if err != '':
                print '[stderr]'
                sys.stdout.write(err)

        expect_eq(0, returncode)
        return returncode, out, err

    @classmethod
    def run_fail(cls, *args, **argd):
        returncode, out, err = cls.run(*args, **argd)
        if returncode == 0:
            if out != '':
                print '[stdout]'
                sys.stdout.write(out)
            if err != '':
                print '[stderr]'
                sys.stdout.write(err)

        expect_ne(0, returncode)
        return returncode, out, err

    @classmethod
    def push_piped(cls, filename, content, **argd):
        return cls.run_ok('push', '-piped', filename, input=content, **argd)

    @classmethod
    def pull_piped(cls, filename, **argd):
        _, out, _ = cls.run_ok('pull', '-piped', filename, **argd)
        return out

    @classmethod
    def trash(cls, *filename, **argd):
        return cls.run_ok('trash', *filename, **argd)

    @classmethod
    def list(cls, path='', recursive=False, **argd):
        extraflags = ['-no-prompt']
        if recursive:
            extraflags += ['-r', '-m=-1']
        _, out, _ = cls.run_ok('list', path, extraflags=extraflags, **argd)
        return sorted(out.splitlines())

    @classmethod
    def erase_all(cls):
        to_trash = []
        for path in cls.list(''):
            assert path[0] == '/' and path[1:]
            to_trash.append(path[1:])
        if to_trash:
            cls.trash(*to_trash, input='y')
            cls.run_ok('emptytrash', '-no-prompt')


@contextmanager
def setup_files(name, *files):
    print '#', name
    try:
        os.unlink(g_driveignore_fn)
    except OSError:
        pass
    for path, content in files:
        Drive.push_piped(path, content)

    try:
        yield
    except Exception:
        if g_fail_stop:
            raise
        traceback.print_exc()

    print '# clean up'
    Drive.erase_all()
    print


def verify_files(*files):
    for path, content in files:
        expect_eq(content, Drive.pull_piped(path))


def test_basic():
    # Most tests depend on these functionality
    fn = 'foo.txt'
    data = 'foobar'

    print '# basic tests'
    Drive.push_piped(fn, data)
    expect_eq(data, Drive.pull_piped(fn))
    Drive.trash(fn, input='y')
    print


def test_list():
    with setup_files('list empty drive'):
        expect_eq([], Drive.list(''))

    with setup_files('list folder',
                     ['a/b/c.txt', 'foobar']):
        expect_eq(['/a'], Drive.list(''))
        expect_eq(['/a/b'], Drive.list('a'))
        expect_eq(['/a/b/c.txt'], Drive.list('a/b'))

    with setup_files('list file, issue #97',
                     ['a/b/c.txt', 'foobar']):
        expect_eq(['/a/b/c.txt'], Drive.list('a/b/c.txt'))

    with setup_files('list not-found, issue #95'):
        _, out, err = Drive.run_fail('list', 'not-found')
        expect_eq('', out)
        expect_ne('', err)


def test_rename():
    with setup_files('rename file in root',
                     ['a.txt', 'a']):
        Drive.run_ok('rename', 'a.txt', 'abc.txt')
        expect_eq(['/abc.txt'], Drive.list())

    with setup_files('rename file in folder',
                     ['b/b.txt', 'b']):
        Drive.run_ok('rename', 'b/b.txt', 'c.txt')
        expect_eq(['/b', '/b/c.txt'], Drive.list(recursive=True))

    # special cases
    with setup_files('rename file to self in root',
                     ['b.txt', 'b']):
        Drive.run_ok('rename', 'b.txt', 'b.txt')
        expect_eq(['/b.txt'], Drive.list(recursive=True))
    with setup_files('rename file to self in folder',
                     ['b/b.txt', 'b']):
        Drive.run_ok('rename', 'b/b.txt', 'b.txt')
        expect_eq(['/b', '/b/b.txt'], Drive.list(recursive=True))

    with setup_files('rename to existing file',
                     ['a.txt', 'a'], ['b.txt', 'b']):
        _, out, err = Drive.run_fail('rename', 'a.txt', 'b.txt')
        expect_true('already exists' in err)
        expect_eq(['/a.txt', '/b.txt'], Drive.list(recursive=True))
        verify_files(['a.txt', 'a'], ['b.txt', 'b'])

    with setup_files('rename special path handling',
                     ['a/b/c.txt', 'c'], ['a/a.txt', 'a']):
        Drive.run_ok('rename', 'a/a.txt', 'b/c.txt')
        expect_eq(['/a', '/a/b', '/a/b/c.txt', '/a/b/c.txt'], Drive.list(recursive=True))


def test_move():
    # basic case
    with setup_files('move folder to another',
                     ['a/a.txt', 'a'], ['b/b.txt', 'b']):
        Drive.run_ok('move', 'a', 'b')
        expect_eq(['/b', '/b/a', '/b/a/a.txt', '/b/b.txt'], Drive.list(recursive=True))

    with setup_files('move multiple files',
                     ['a/a.txt', 'a'], ['b/b.txt', 'b'], ['c/c.txt', 'c']):
        Drive.run_ok('move', 'a/a.txt', 'b/b.txt', 'c')
        expect_eq(['/a', '/b', '/c', '/c/a.txt', '/c/b.txt', '/c/c.txt'], Drive.list(recursive=True))

        Drive.run_ok('move', 'c/a.txt', 'c/b.txt', 'c/c.txt', '')
        expect_eq(['/a', '/a.txt', '/b', '/b.txt', '/c', '/c.txt'], Drive.list(recursive=True))

    with setup_files('move file to file',
                     ['a.txt', 'a'], ['b.txt', 'b']):
        Drive.run_fail('move', 'a.txt', 'b.txt')
        expect_eq(['/a.txt', '/b.txt'], Drive.list(recursive=True))
        verify_files(['a.txt', 'a'], ['b.txt', 'b'])

    # special cases
    with setup_files('move file to the same folder',
                     ['a/b.txt', 'b']):
        Drive.run_ok('move', 'a/b.txt', 'a')
        expect_eq(['/a', '/a/b.txt'], Drive.list(recursive=True))

    with setup_files('move folder to its parent',
                     ['a/b/c.txt', 'c']):
        Drive.run_ok('move', 'a/b', 'a')
        expect_eq(['/a', '/a/b', '/a/b/c.txt'], Drive.list(recursive=True))

    with setup_files('move folder to its child',
                     ['a/b/c.txt', 'c']):
        Drive.run_fail('move', 'a', 'a/b')
        expect_eq(['/a', '/a/b', '/a/b/c.txt'], Drive.list(recursive=True))

    with setup_files('move multiple files and duplicated',
                     ['a/foo.txt', 'a'], ['b/foo.txt', 'b'], ['c/c.txt', 'c']):
        _, _, err = Drive.run_ok('move', 'a/foo.txt', 'b/foo.txt', 'c')
        expect_ne('', err)
        expect_eq(['/a', '/b', '/b/foo.txt', '/c', '/c/c.txt', '/c/foo.txt'], Drive.list(recursive=True))


def test_stat():
    cases = [
        '',
        'foobar',
        ''.join(map(chr, range(256))),
    ]
    for data in cases:
        print 'try', repr(data)
        with setup_files('stat file with size=%d' % len(data),
                         ['foo.txt', data]):
            _, out, _ = Drive.run_ok('stat', 'foo.txt')
            expect_true(re.search(r'Bytes\s+%s' % len(data), out))
            expect_true(re.search(r'DirType\s+file', out))
            expect_true(re.search(r'MimeType\s+text/plain', out))
            expect_true(re.search(r'Md5Checksum\s+%s' % hashlib.md5(data).hexdigest(), out))


def test_pull():
    with setup_files('pull -piped not-found file, issue #95'):
        _, out, err = Drive.run_fail('pull', '-piped', 'not-found')
        expect_eq('', out)
        expect_ne('', err)

    with setup_files('pull -piped folder',
                     ['a/a.txt', '']):
        _, out, err = Drive.run_fail('pull', '-piped', 'a')
        expect_eq('', out)
        expect_ne('', err)


def test_trash():
    with setup_files('trash file',
                     ['a.txt', 'a']):
        Drive.trash('a.txt', input='y')
        expect_eq([], Drive.list())

    with setup_files('trash folder',
                     ['a/b.txt', 'b']):
        Drive.trash('a/b.txt', input='y')
        expect_eq(['/a'], Drive.list(recursive=True))
        Drive.trash('a', input='y')
        expect_eq([], Drive.list())

    with setup_files('trash multiple files',
                     ['a.txt', ''], ['b.txt', ''], ['c.txt', '']):
        _, _, err = Drive.run_ok('trash', 'a.txt', 'b.txt', 'c.txt', input='y')
        expect_eq([], Drive.list())

    with setup_files('trash non-existing file'):
        _, _, err = Drive.run_fail('trash', 'not-found', input='y')
        expect_ne('', err)


def main():
    init()

    test_basic()

    test_list()
    test_rename()
    test_move()
    test_stat()
    test_pull()
    test_trash()

    print 'ok', count_ok
    print 'bad', count_bad


if __name__ == '__main__':
    main()
