# Simple FTP Server

## SenseTime Container Platform Tranining Program

A simple FTP server written in golang.

It implements several simple instructions, which can enable basic usage for ftp clients.

## Version

    v1

## Deployment

Create a working directory of go (e.g., /go), create folders /go/bin, /go/src, and /go/src/myftp.
Put all files into the directory /go/src/myftp.

## Compile and Install

Get into the "myftp" directory in the working directory.

    $ cd /go/src/myftp
    $ go build
    $ go install myftp

## Run

Run in native system:

    $ cd /go/bin/myftp
    $ ./myftp -native

Run with parameters (port, host, directory)

    $ cd /go/bin/myftp
    $ ./myftp -native -p xxxx -a xxx.xxx.xxx.xxx -d /xx/xx

Get help message

    $ /go/bin/myftp -h
