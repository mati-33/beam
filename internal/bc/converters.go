package bc

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func ipv4ToBinStr(ip net.IP) string {
	parts := make([]string, 4)
	for i, o := range ip {
		parts[i] = fmt.Sprintf("%08b", o)
	}
	return strings.Join(parts, "")
}

func binToHexStr(bin string) (string, error) {
	n, err := strconv.ParseUint(bin, 2, 64)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", n), nil
}

func hexToBinStr(hex string) (string, error) {
	n, err := strconv.ParseUint(hex, 16, 64)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%b", n), nil
}

func binaryIPv4ToDecimal(bin string) (string, error) {
	if len(bin) != 32 {
		return "", errors.New("input must be a 32-bit binary string")
	}

	parts := make([]string, 4)
	for i := range 4 {
		byteStr := bin[i*8 : (i+1)*8]

		val, err := strconv.ParseUint(byteStr, 2, 8)
		if err != nil {
			return "", fmt.Errorf("invalid binary byte '%s': %v", byteStr, err)
		}

		parts[i] = fmt.Sprintf("%d", val)
	}

	return strings.Join(parts, "."), nil
}
