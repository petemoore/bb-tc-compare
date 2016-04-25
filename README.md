Compare Windows Buildbot builds to Taskcluster builds
=====================================================

Long story short, this is a utility script written for the explicit purpose of trying to debug differences between Buildbot Windows builds of Firefox with TaskCluster Windows builds of Firefox.

It is not meant to be of for a broad number of people, really just the people helping out on this project.

Requirements
------------

Have go 1.5/1.6 installed on your system, and your GOPATH configured. May also work with earlier versions.

Installation
------------

```
go get -u github.com/petemoore/bb-tc-compare
```

Running
-------

```
"${GOPATH}/bin/bb-tc-compare" <user>
```

for example:

```
bb-tc-compare pmoore@mozilla.com
```

This will create a directory structure under your current working directory, e.g.

```
$ find 'pmoore@mozilla.com' -type f
pmoore@mozilla.com/1e2f5b13b15fcc8b2502766d321bc435f3c7606e/opt/windowsxp/tc
pmoore@mozilla.com/2e9fa4a3adccd30f87965ce0c7b46380884b22a5/opt/windowsxp/bb
pmoore@mozilla.com/2f2fe1ba78aa07359b41754d592c1b720c316f64/opt/windowsxp/bb
pmoore@mozilla.com/3542014e3539d165c11da0fc0f8f0d30f2f36ec7/opt/windowsxp/bb
pmoore@mozilla.com/49bb5b12bb6c74970be0f883d037fd98447fa173/opt/windowsxp/bb
pmoore@mozilla.com/49bb5b12bb6c74970be0f883d037fd98447fa173/opt/windowsxp/tc
pmoore@mozilla.com/c82553993215234d6c8c1d93772a863603a99ea6/opt/windowsxp/bb
pmoore@mozilla.com/c82553993215234d6c8c1d93772a863603a99ea6/opt/windowsxp/tc
pmoore@mozilla.com/e30210718ae95950a43829da22243a8f52299fcd/opt/windowsxp/tc
pmoore@mozilla.com/f26427f7ee0f62abfd04de379b17f3a6af1c8363/opt/windowsxp/bb
```

The idea is then you can use a diff tool, like `vim -d` to compare bb builds with tc builds.

The `bb` and `tc` files are the log files for the mozharness steps of the Firefox Windows builds. They have been "normalised" to remove timestamps, and references to e.g. the build directory. The idea is that differences like these get filtered out, so only significant differences remain.

Note, this is a work-in-progress, and can be hacked on by anyone helping out with analysing the differences.

If you don't want to compile code between runs, simply use `go run bb-tc-compare.go` from where you have it checked out on your filesystem. Please feed changes back to me via a PR, as they will be useful to me!
