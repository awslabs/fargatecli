package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var validProtocol = regexp.MustCompile("(?i)\\ATCP|HTTP(S)?\\z")

func inflatePort(portExpr string) (Port, error) {
	switch {
	case portExpr == "80":
		return Port{Protocol: "HTTP", Port: 80}, nil
	case portExpr == "443":
		return Port{Protocol: "HTTPS", Port: 443}, nil
	case strings.Index(portExpr, ":") > 1:
		parts := strings.Split(portExpr, ":")
		protocol := strings.ToUpper(parts[0])
		port, err := strconv.ParseInt(parts[1], 10, 64)

		if err != nil {
			return Port{}, err
		}

		return Port{Protocol: protocol, Port: port}, nil
	default:
		port, err := strconv.ParseInt(portExpr, 10, 64)

		if err != nil {
			return Port{}, err
		}

		return Port{Protocol: "TCP", Port: port}, nil
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
		errs = append(errs, fmt.Errorf("Invalid protocol %s [specify TCP, HTTP, or HTTPS]", port.Protocol))
	}

	if port.Port < 1 || port.Port > 65535 {
		errs = append(errs, fmt.Errorf("Invalid port %d [specify within 1 - 65535]", port.Port))
	}

	return errs
}
