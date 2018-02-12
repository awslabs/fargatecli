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
		return Port{80, "HTTP"}, nil
	case portExpr == "443":
		return Port{443, "HTTPS"}, nil
	case strings.Index(portExpr, ":") > 1:
		parts := strings.Split(portExpr, ":")
		protocol := strings.ToUpper(parts[0])
		number, err := strconv.ParseInt(parts[1], 10, 64)

		if err != nil {
			return Port{}, err
		}

		return Port{number, protocol}, nil
	default:
		port, err := strconv.ParseInt(portExpr, 10, 64)

		if err != nil {
			return Port{}, err
		}

		return Port{port, "TCP"}, nil
	}
}

func inflatePorts(portExprs []string) ([]Port, []error) {
	var ports []Port
	var errs []error

	for _, portExpr := range portExprs {
		port, err := inflatePort(portExpr)

		if err != nil {
			errs = append(errs, err)
		}

		ports = append(ports, port)
	}

	return ports, errs
}

func validatePort(port Port) []error {
	var errs []error

	if !validProtocol.MatchString(port.Protocol) {
		errs = append(errs, fmt.Errorf("Invalid protocol %s (specify TCP, HTTP, or HTTPS)", port.Protocol))
	}

	if port.Number < 1 || port.Number > 65535 {
		errs = append(errs, fmt.Errorf("Invalid port %d (specify within 1 - 65535)", port.Number))
	}

	return errs
}