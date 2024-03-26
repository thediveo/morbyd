/*
Package net provides options to configure new Docker (custom) networks.

Subpackages provide further driver-specific options:
  - [github.com/thediveo/morbyd/net/bridge] for bridge-related options.
  - [github.com/thediveo/morbyd/net/macvlan] for MACVLAN-related options.
  - [github.com/thediveo/morbyd/net/ipvlan] for IPVLAN-related options.

# Usage

To create a new internal custom Docker network, including IPAM pool
configuration and bridge-specific configuration:

	sess.CreateNetwork(ctx, "my-custom-notwork",
	    net.WithInternal(),
	    bridge.WithBridgeName("brrr-42"),
	    net.WithIPAM(ipam.WithPool("0.0.1.0/24", ipam.WithRange("0.0.1.16/28"))),
	)
*/
package net
