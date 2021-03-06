/*
*    mcast - Command line tool and library for testing multicast traffic
*    flows and stress testing networks and devices.
*    Copyright (C) 2018 Will Smith
*
*    This program is free software: you can redistribute it and/or modify
*    it under the terms of the GNU General Public License as published by
*    the Free Software Foundation, either version 3 of the License, or
*    (at your option) any later version.
*
*    This program is distributed in the hope that it will be useful,
*    but WITHOUT ANY WARRANTY; without even the implied warranty of
*    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
*    GNU General Public License for more details.
*
*    You should have received a copy of the GNU General Public License
*    along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package multicast

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
)

// ComputeChecksum returns the 16bit 1's compliment for the given byte slice
func ComputeChecksum(buf []byte) uint16 {
	sum := uint32(0)

	for ; len(buf) >= 2; buf = buf[2:] {
		sum += uint32(buf[0])<<8 | uint32(buf[1])
	}
	if len(buf) > 0 {
		sum += uint32(buf[0]) << 8
	}
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	csum := ^uint16(sum)
	/*
	 * From RFC 768:
	 * If the computed checksum is zero, it is transmitted as all ones (the
	 * equivalent in one's complement arithmetic). An all zero transmitted
	 * checksum value means that the transmitter generated no checksum (for
	 * debugging or for higher level protocols that don't care).
	 */
	if csum == 0 {
		csum = 0xffff
	}
	return csum
}

// ComputeChecksumBytes returns the 16bit checksum split into high and low bytes for a given byte slice.
func ComputeChecksumBytes(buf []byte) (byte, byte) {
	checksum := ComputeChecksum(buf)
	b1 := byte(checksum >> 8)   //byte(0xee) // byte(checksum & 0x00FF)
	b2 := byte(checksum & 0xFF) //byte(0x9b)          // byte((checksum & 0xFF00) >> 1)
	return b1, b2
}

// IP4ToInt returns a 32bit integer representation for a given ipv4 address.
func IP4ToInt(ip net.IP) uint32 {
	ip4 := []byte(ip.To4())
	return uint32((uint32(ip4[0]) << 24) | (uint32(ip4[1]) << 16) | (uint32(ip4[2]) << 8) | uint32(ip4[3]))
}

// IntToIP4 returns a net.IP for a given 32bit integer representing an ip address.
func IntToIP4(ipInt uint32) net.IP {
	return net.IPv4(
		byte(ipInt>>24),
		byte(ipInt>>16),
		byte(ipInt>>8),
		byte(ipInt),
	)
}

// IPList will return a slice of net.IP addresses for the given ip and mask.
// The returned slice will be based on the network that the provided ip falls
// in and not necessarily the exact ip given. The returned slice will also
// include both the network address and the broadcast address as the first
// and last items of the slice.
func IPList(network string, mask int) ([]net.IP, error) {
	_, ipnet, err := net.ParseCIDR(fmt.Sprintf("%v/%d", network, mask))
	if err != nil {
		return nil, err
	}
	networkBits, totalBits := ipnet.Mask.Size()
	hostBits := totalBits - networkBits
	numberOfHosts := uint32(math.Pow(float64(2), float64(hostBits)))

	hostAddresses := make([]net.IP, numberOfHosts)
	networkInt := IP4ToInt(ipnet.IP)
	for i := uint32(0); i < numberOfHosts; i++ {
		hostAddresses[i] = IntToIP4(networkInt | i)
	}

	return hostAddresses, nil
}

// SplitCIDR returns the ip, or network portion and the mask as 2 separate values for a given address.
func SplitCIDR(address string) (string, int, error) {
	if !strings.Contains(address, "/") {
		return address, 32, nil
	}
	addressParts := strings.Split(address, "/")
	network := addressParts[0]
	mask, err := strconv.ParseInt(addressParts[1], 10, 32)
	if err != nil {
		return "", 0, err
	}
	return network, int(mask), nil
}

// IPListCIDR is a convenience function combining SplitCIDR and IPList.
// It takes an address in CIDR format and returns a slice of IPs within that range.
func IPListCIDR(address string) ([]net.IP, error) {
	network, mask, err := SplitCIDR(address)
	if err != nil {
		return nil, err
	}
	return IPList(network, mask)
}

// GetInterface returns the interface associated with the name provided.
// It wraps the net.InterfaceByName only adding functionality to allow
// the specified interface to be an empty string. This helps with command
// line processing where the default value is an empty string.
func GetInterface(interfaceName string) (*net.Interface, error) {
	var localInterface *net.Interface
	var err error
	if interfaceName != "" {
		localInterface, err = net.InterfaceByName(interfaceName)
	}
	return localInterface, err
}
