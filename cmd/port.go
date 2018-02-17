package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Port struct {
	Number   int64
	Protocol string
}

func (p *Port) Empty() bool {
	return p.Number == 0 && p.Protocol == ""
}

func (p *Port) String() string {
	return fmt.Sprintf("%s:%d", p.Protocol, p.Number)
}

var validProtocol = regexp.MustCompile("(?i)\\ATCP|HTTP(S)?\\z")

func inflatePort(portExpr string) (Port, error) {
	switch {
	case portExpr == "80":
		return buildPort(portExpr, "HTTP")
	case portExpr == "443":
		return buildPort(portExpr, "HTTPS")
	case strings.Index(portExpr, ":") > 1:
		parts := strings.Split(portExpr, ":")
		protocol, number := strings.ToUpper(parts[0]), parts[1]

		return buildPort(number, protocol)
	default:
		return buildPort(portExpr, "TCP")
	}
}

func buildPort(inputNumber, inputProtocol string) (Port, error) {
	number, err := strconv.ParseInt(inputNumber, 10, 64)

	if err != nil {
		if _, ok := err.(*strconv.NumError); ok {
			return Port{}, fmt.Errorf("could not parse port number from %s", inputNumber)
		} else {
			return Port{}, err
		}
	}

	return Port{number, inputProtocol}, nil
}

func inflatePorts(portExprs []string) ([]Port, []error) {
	var ports []Port
	var errs []error

	for _, portExpr := range portExprs {
		if port, err := inflatePort(portExpr); err == nil {
			ports = append(ports, port)
		} else {
			errs = append(errs, err)
		}
	}

	return ports, errs
}

func validatePort(port Port) (errs []error) {
	if !validProtocol.MatchString(port.Protocol) {
		errs = append(errs, fmt.Errorf("invalid protocol %s (specify TCP, HTTP, or HTTPS)", port.Protocol))
	}

	if port.Number < 1 || port.Number > 65535 {
		errs = append(errs, fmt.Errorf("invalid port %d (specify within 1 - 65535)", port.Number))
	}

	return
}
