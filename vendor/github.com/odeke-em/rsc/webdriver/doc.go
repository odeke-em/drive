// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package webdriver implements a client for the Selenium WebDriver wire protocol.
// It can be used with the ChromeDriver server and perhaps with the Selenium server.
//
// This package implements the wire protocol described at
// http://code.google.com/p/selenium/wiki/JsonWireProtocol.
//
// For Selenium, see http://code.google.com/p/selenium/.
// For ChromeDriver, see http://code.google.com/p/chromedriver/.
//
// There is an ongoing W3C standardization effort for WebDriver at
// https://dvcs.w3.org/hg/webdriver/raw-file/tip/webdriver-spec.html.
// This package may or may not implement that protocol, depending on
// how close it is to the one linked above.
//
// INCOMPLETE AND UNDOCUMENTED.
//
package webdriver
