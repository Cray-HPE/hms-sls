# SLS Overview

## What is currently stored in SLS?

## Hardware information

* Source of truth for HMS Serviers (MEDS, REDS, HMS-Discovery, and HSM) when trying to identify and discovery new hardware in the system
* Partial topology of the HMN network. SLS currently has cabling information about how the BMCs on the River HMN are connected . BMC <-> Leaf (CSM 1.0)/Leaf BMC (CSM 1.2+) Management switches.
* Role, Subrole, NID, and Alias information for nodes (Management, Compute, Application).
* Management switches with their aliases
* SLS currently has the following hardware types populated in it:

    * Cabinet
    * ChassisBMC – Hill/Mountain only
    * CabinetPDUController
    * RouterBMC
    * MgmtSwitchConnector – River Only
    * MgmtSwitch – River Only
    * MgmtHLSwitch – River Only
    * MgmtCDUSwitch – Mountain/Hill
    * Node

### Network Information

* SLS is the authority for the prescribed layout of management networking on the system.
* A SLS Network is…
  * A logical concept. We have allocated a CIDR/IP Address range for a specific purpose. Examples include the 
     * Node Management Network (NMN),
     * Hardware Management Network (HMN)
     * Customer Access Network (CAN)
  * A network is broken into smaller subnets (/22) that fit within the networks CIDR (/17)
* A SLS Subnet is…
  * Can contain a DHCP Address pool
  * VLAN Information
  * IP Reservations: IP Address, Name, and Aliases
  * Examples include:
    * HMN Management Network Infrastructure
    * HMN Bootstrap DHCP Subnet
    * NMN UAIs

## What services interact with SLS?

* hms-discovery cronjob
  * River HMN Cabling information to identify unknown hardware
  * Listing of Management Switches in the System to query via SNMP
* REDS
  * Listing of Management NCNs and Columbia Switches
* MEDS
  * Looks to SLS for what Mountain/Hill cabinets are in the system
* Hardware State Manager (HSM)
  * Role, Subrole, and NID information
* KEA – DHCP
  * DHCP Subnet ranges
* Unbound – DNS
  * IP Reservations to populate static DNS records. Like packages.local, registry.local, and api-gw-service-nmn.local
  * Node Alias information. For example for the node x3000c0s19b1n0, look to SLS to create a DNS entry for the friendly name automatically ncn-000001-nmn.
* PowerDNS – DNS
  * IP Reservations to populate DNS
* CANU

![sls-overview.svg](../docs/images/sls_overview.svg)
