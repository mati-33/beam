package bc

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
)

func getLocalIPv4() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	localIPv4 := localAddr.IP.To4()
	if localIPv4 == nil {
		return nil, errors.New("failed to obtaing local IPv4")
	}

	return localAddr.IP, nil
}

func getNetworkPrefix() (int, error) {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {

			ip, ipNet, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			if !ip.IsPrivate() {
				continue
			}

			prefix, _ := ipNet.Mask.Size()
			return prefix, nil
		}
	}

	return 0, errors.New("failed to find network prefix")
}

func getHostBits(ip net.IP, npref int) string {
	ipBinary := ipv4ToBinStr(ip)
	return ipBinary[npref:]
}

func randomNumRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func BeamCodeHex() string {
	return fmt.Sprintf("%02x", randomNumRange(1, 255))
}

func HostBitsHex() (string, error) {
	localIPv4, err := getLocalIPv4()
	if err != nil {
		return "", err
	}

	npref, err := getNetworkPrefix()
	if err != nil {
		return ipv4ToHexStr(localIPv4), nil
	}

	hostBits := getHostBits(localIPv4, npref)
	hostBitsHex, err := binToHexStr(hostBits)
	if err != nil {
		return "", nil
	}

	return hostBitsHex, nil
}

func AbsorberAddress(hostBitsHex string) (string, error) {
	hostBitsBin, err := hexToBinStr(hostBitsHex)
	if err != nil {
		return "", err
	}

	localIPv4, err := getLocalIPv4()
	if err != nil {
		return "", err
	}

	localIPv4Bin := ipv4ToBinStr(localIPv4)
	addressBin := localIPv4Bin[:len(localIPv4Bin)-len(hostBitsBin)] + hostBitsBin

	address, err := binaryIPv4ToDecimal(addressBin)
	if err != nil {
		return "", err
	}

	return address, nil
}
