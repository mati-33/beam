package bc

import (
	"net"
	"testing"
)

func Test_ipv4ToBinStr(t *testing.T) {
	tests := []struct {
		name string
		ip   net.IP
		want string
	}{
		{
			"ipv4 1",
			net.IPv4(192, 168, 1, 14).To4(),
			"11000000101010000000000100001110",
		},
		{
			"ipv4 2",
			net.IPv4(172, 16, 2, 211).To4(),
			"10101100000100000000001011010011",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ipv4ToBinStr(tt.ip)
			if got != tt.want {
				t.Errorf("ipv4ToBinStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_binToHexStr(t *testing.T) {
	tests := []struct {
		name    string
		bin     string
		want    string
		wantErr bool
	}{
		{
			"255",
			"11111111",
			"ff",
			false,
		},
		{
			"1",
			"1",
			"1",
			false,
		},
		{
			"invalid binary",
			"12",
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := binToHexStr(tt.bin)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("binToHexStr() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("binToHexStr() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("binToHexStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hexToBinStr(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		want    string
		wantErr bool
	}{
		{
			"255",
			"ff",
			"11111111",
			false,
		},
		{
			"1",
			"1",
			"1",
			false,
		},
		{
			"invalid hex",
			"zz",
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := hexToBinStr(tt.hex)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("hexToBinStr() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("hexToBinStr() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("hexToBinStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_binaryIPv4ToDecimal(t *testing.T) {
	tests := []struct {
		name    string
		bin     string
		want    string
		wantErr bool
	}{
		{
			"to short",
			"100010010",
			"",
			true,
		},
		{
			"not binary",
			"hello world",
			"",
			true,
		},
		{
			"ipv4 1",
			"11000000101010000000000100001110",
			"192.168.1.14",
			false,
		},
		{
			"ipv4 2",
			"10101100000100000000001011010011",
			"172.16.2.211",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := binaryIPv4ToDecimal(tt.bin)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("binaryIPv4ToDecimal() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("binaryIPv4ToDecimal() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("binaryIPv4ToDecimal() = %v, want %v", got, tt.want)
			}
		})
	}
}
