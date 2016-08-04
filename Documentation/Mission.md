# XDOJ 3.x Development Mission Statement

XDOJ 3.x development is a project of Laboratory of ACM/ICPC,
Xidian University. The purpose of XDOJ 3.x is to replace currently
used XDOJ 2.x (based on [HUSTOJ](https://github.com/zhblue/hustoj))
system.

## Objections

### High Performance

We want XDOJ 3.x to handle 2000 judgements and 50000 other HTTP
requests in 10 min smoothly (this happens at the beginning of a
contest).

### High Reliability

We want no misjudgement, and minimalized number of `Judgement Error`s.
XDOJ-v3 should never forget to judge any submittions.

### Easy Deployment

We want XDOJ 3.x to be easily deployed on all i686 or x86-64 linux
platforms. Source code tarball, binary package, and images (Docker
or QEMU) will be avalible with detailed deployment documentation.

### Easy Administration

We want OJ administrators and problem setters to manage XDOJ 3.x
easily.

### Enhanced Safety

We have to ensure no submittions can rend XDOJ 3.x unuseable intentionally
or not intentionally.

## High Priority Components

`xdoj3-judged` would be jury of XDOJ 3.x. It compiles, runs, traces
submittions, and check the output of them.

`xdoj3-httpd` would be HTTP server of XDOJ 3.x. It handles HTTP requests
from Web browser, to interact with participants and others.

## Long-Range Goals

`xdoj3-sim` would calculate source code similarity, in order to help
jurys to root out cheaters.

`xdoj3-spj` would be a library for problem setters to write special judges
easily.
