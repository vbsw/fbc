/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import "strconv"

func messageShortInfo() string {
	return "Run 'fbc --help' for usage."
}

func messageHelp() string {
	message := "\nUSAGE\n"
	message += "  fbc (INFO | ( COMMAND INPUT-DIR {OUTPUT-DIR FILTER OPTION} ))\n\n"
	message += "INFO\n"
	message += "  -h, --help       print this help\n"
	message += "  -v, --version    print version\n"
	message += "  -e, --example    print example\n"
	message += "  -c, --copyright  print copyright\n"
	message += "COMMAND\n"
	message += "  count            count files\n"
	message += "  cp               copy files\n"
	message += "  mv               move files\n"
	message += "  print            print file names\n"
	message += "  rm               delete files\n"
	message += "OPTION\n"
	message += "  -o, --or         filter is OR (not AND)\n"
	message += "  -r, --recursive  recursive file iteration"
	return message
}

func messageVersion() string {
	return "1.0.0"
}

func messageExample() string {
	message := "\nEXAMPLES\n"
	message += "   fbc cp ./ ../bak bob alice\n"
	message += "   fbc mv ./*.txt ../bak bob alice\n"
	message += "   fbc rm ./*.txt bob alice"
	return message
}

func messageCopyright() string {
	message := "Copyright 2020, Vitali Baumtrok (vbsw@mailbox.org).\n"
	message += "Distributed under the Boost Software License, Version 1.0."
	return message
}

func messageFinished(count int) string {
	var fileStr string
	if count == 1 {
		fileStr = " file"
	} else {
		fileStr = " files"
	}
	return "finished: " + strconv.Itoa(count) + fileStr
}

func messageError(err error) string {
	return "error: " + err.Error()
}
