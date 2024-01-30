/*
Package ipam provides options for IP address management (“IPAM”), making
Docker's IPAM-related API data structures more accessible.

# Usage

The IPAM configuration options can be used in the context of creating a new
custom Docker network as follows:

	morbyd.NewNetwork(ctx, "my-custom-notwork",
	    net.WithIPAM(ipam.WithPool("0.0.1.0/24", ipam.WithRange("0.0.1.16/28"))),
	)

Please note using [github.com/thediveo/morbyd/net.WithIPAM] for configuring an
IPAM driver, including setting up multiple address pools by way of [WithPool].

# References

For further background information on Docker IPAMs, please see:

  - [IPAM Drivers] (Moby libnetwork documentation)
  - [builtin IPAM drivers] (Moby libnetwork codebase)

[IPAM Drivers]: https://github.com/moby/libnetwork/blob/master/docs/ipam.md
[builtin IPAM drivers]: https://github.com/moby/libnetwork/tree/master/ipams
*/
package ipam
