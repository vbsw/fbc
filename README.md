# fbc

[![GoDoc](https://godoc.org/github.com/vbsw/fbc?status.svg)](https://godoc.org/github.com/vbsw/fbc) [![Go Report Card](https://goreportcard.com/badge/github.com/vbsw/fbc)](https://goreportcard.com/report/github.com/vbsw/fbc) [![Stability: Experimental](https://masterminds.github.io/stability/experimental.svg)](https://masterminds.github.io/stability/experimental.html)

## About
fbc (file by content) allows various commands on files filtered by their content. fbc is published on <https://github.com/vbsw/fbc> and <https://gitlab.com/vbsw/fbc>.

Download [binaries](https://github.com/vbsw/fbc/archive/bin.zip).

## Copyright
Copyright 2020, 2021, Vitali Baumtrok (vbsw@mailbox.org).

fbc is distributed under the Boost Software License, version 1.0. (See accompanying file LICENSE or copy at http://www.boost.org/LICENSE_1_0.txt)

fbc is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the Boost Software License for more details.

## Usage

	fbc (INFO | ( COMMAND INPUT-DIR {OUTPUT-DIR FILTER OPTION} ))

	INFO
		-h, --help        print this help
		-v, --version     print version
		-e, --example     print example
		-c, --copyright   print copyright
	COMMAND
		count             count files
		cp                copy files
		mv                move files
		print             print file names
		rm                delete files
	OPTION
		-o, --or          filter is OR (not AND)
		-r, --recursive   recursive file iteration

## Examples

Copy any file containing the words "alice" and "bob"

	$ fbc cp ./ ../bak alice bob

Move text files containing the words "alice" and "bob"

	$ fbc mv "./*.txt" ../bak alice bob

Delete text files containing the words "alice" and "bob"

	$ fbc rm "./*.txt" alice bob

## References
- https://golang.org/doc/install
- https://git-scm.com/book/en/v2/Getting-Started-Installing-Git
