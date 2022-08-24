1.  [CASMHMS](index.html)
2.  [CASMHMS Home](CASMHMS-Home_119901124.html)
3.  [Design Documents](Design-Documents_127906417.html)
4.  [Discovery](Discovery_227479292.html)

# <span id="title-text"> CASMHMS : System Layout Service (SLS) Design Documentation </span>

Created by <span class="author"> Steven Presser</span>, last modified by
<span class="editor"> Matt Kelly</span> on Oct 29, 2019

The System Layout Service (SLS) holds information about a Cray Shasta
system "as configured".  That is, it contains information about how the
system was designed.  SLS reflects the system if all hardware were
present and working.  SLS does not keep track of hardware state or
identifiers; for that type of information, see Hardware State Manager
(HSM).

# Design Principles

SLS is designed with the following principles and requirements in mind:

-   Flexibility.  Our customers are very specific about how their
    systems are constructed.  Additionally, the hardware involved in
    Shasta is expected to shift over time, and not always in a manner
    that we can predict.  As few assumptions as possible should be made
    about the hardware.
-   Ease of use.  The API should follow RESTful principles while
    maintaining as much ease of use as possible.

# Object Model

As everything in SLS fundamentally represents something that exists in
the real world (or could exist in the real world), it makes sense to
deal with everything as objects.  The [Shasta HSS Component Naming
Convention](https://connect.us.cray.com/confluence/display/HSOS/Shasta+HSS+Component+Naming+Convention)
page and the
<a href="https://stash.us.cray.com/projects/HMS/repos/hms-common/browse/pkg/base/hmstypes.go#32" class="external-link">hmstypes.go</a>
file include an exhaustive list of hardware types present in a Shasta
system.

## Base Functions and Properties

These are the properties available on every object in SLS:

-   Parent – Each object in SLS has a parent object.  This gets a
    reference to that object.
-   Children – Objects may have children.  This gets references to the
    children
-   Xname – Every object has an xname – a unique identifier for that
    object.  This gets the xname as a string
-   Type – a hardware type as listed in the "Component Enumeration"
    table in [Shasta HSS Component Naming
    Convention](https://connect.us.cray.com/confluence/display/HSOS/Shasta+HSS+Component+Naming+Convention)
-   Class – what kind of hardware this is.  Either "river" or "mountain"
-   <span class="inline-comment-marker"
    ref="baed197b-c39b-4f47-afc6-6890cadc2aa9">TypeString</span> – a
    human readable type, as listed
    in <a href="https://stash.us.cray.com/projects/HMS/repos/hms-common/browse/pkg/base/hmstypes.go#32" class="external-link">hmstypes.go</a>

Example:

    "x0c0s0b0" {
            "Parent": "x0c0s0",
            "Children: [
                "x0c0s0b0n0",
                "x0c0s1b0n0",
                "x0c0r0b0n0",
                "x0c0r1b0n0",
                ...
            ],
            "XName": "x0c0s0b0",
            "Type": "comptype_ncard",
            "Class": "mountain",
        }

  

## Object-specific Properties

In addition to the base functions and properties, the following items
have these additional or modified properties:

  

### cabinet

Represents a cabinet.  This is typically a Mountain cabinet but can also
be used for a river rack.   Mountain cabinets are important to have in
SLS because MEDS uses this data to calculate endpoint information.

    "x1000": {
        "Parent": "s0",
        "Children: [
            "x1000c0",
            "x1000c1",
            ...
        ],
        "XName":"x1000",
        "Type": "comptype_cabinet",
        "TypeString": "Cabinet",
        "Class": "mountain",
        "ExtraProperties": {
            "Network": "HMN",
            "IP6Prefix": "fd66:0:0:0",
            "IP4Base": "10.1.2.100/26",
            "MACprefix": "02"
        }
    }

  

### \*\_pwr\_connector

These items represent a power connection.  They therefore have at least
two ends.

-   Parent – is the comptype\_cab\_pdu this cable is connected to
-   PoweredBy– is the hardware this cable is connected to.  May be any
    type of object.

  

    # River TOR power connector
    "x0c0r0v0" {
        "Parent": "x0c0r0",
        "XName": "x0c0r0v0",
        "Type": "comptype_rtrmod_power_connector",
        "TypeString": "RouterPowerConnector", ## ZZZZ add to hmstypes.go
        "Class": "river"
        "ExtraProperties": {
            "PoweredBy": "x0m0p0v1"  # what PDU socket this power connector is connected to
        }
    }   

### comptype\_hsn\_connector

-   Parent – the switch which this cable is connected to
-   NodeNics – an array of comptype\_hsn\_connector\_ports that this
    cable is connected to

  

    # HSN connector, River or Mountain
    "x0c0r0j0" {
        "Parent": "x0c0r0",
        "XName": "x0c0r0j0",
        "Type": "comptype_hsn_connector",
        "Class": "river",
        "TypeString": "HSNConnector",
        "ExtraProperties: {
            "NodeNics": [        #Node NICs this HSN port connects to
                "x0c0s0b0n0h0",
                "x0c0s0b0n1h1",
                ...
            ]
        }
    }

  

### comptype\_mgmt\_switch

-   IP6addr – an address, or "DHCPv6".  May be omitted; if so, assumed
    to be DHCPv6
-   IP4addr – an address, or "DHCP".  May be omitted; if so, assumed to
    be DHCP
-   <span class="inline-comment-marker"
    ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">Username – the user name
    that should be used when accessing the device</span>
-   <span class="inline-comment-marker"
    ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">Password – the associated
    password</span>
-   <span class="inline-comment-marker"
    ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">SNMPUsername – the
    username that should be used when accessing the SNMP
    interface</span>
-   <span class="inline-comment-marker"
    ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">SNMPAuthPassword – the
    SNMP authorization password associated with the SNMPUsername</span>
-   <span class="inline-comment-marker"
    ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">SNMPAuthProtocol – the
    SNMP authorization protocol</span>
-   <span class="inline-comment-marker"
    ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">SNMPPrivPassword – the
    SNMP privacy password associated with the SNMPUsername</span>
-   <span class="inline-comment-marker"
    ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">SNMPPrivProtocol – the
    SNMP privacy protocol</span>
-   <span class="inline-comment-marker"
    ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">Model – Information on
    the model of the switch.   
    </span>

<span class="inline-comment-marker"
ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">Note that all passwords are
stored in secured storage and will not be visible unless encrypted via
the /dumpstate API.  
</span>

<span class="inline-comment-marker"
ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">  
</span>

    "x0c0w0" {
        "Parent": "x0c0",
        "Children: [
            "x0c0w0j1",
            "x0c0w0j2",
            ...
        ]
        "XName": "x0c0w0",
        "Type": "comptype_mgmt_switch",
        "TypeString": "MgmtSwitch",
        "Class": "river",
        "ExtraProperties": {
            "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
            "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
            "Username": "user_name",  # to access mgmt interface
            "Password": "vault://tok",# ptr to vault URL to get password
            "SNMPUsername": "username",
            "SNMPAuthPassword": "vault://tok",
            "SNMPAuthProtocol": "MD5",
            "SNMPPrivPassword": "vault://tok",
            "SNMPPrivProtocol": "DES",
            "Model": "Dell S3048-ON"
        }
    }

<span class="inline-comment-marker"
ref="38cf2e76-3fb4-44aa-8c8f-f79e8f9658ca">  
</span>

### comptype\_mgmt\_switch\_connector

-   Parent – the switch which this cable is connected to
-   NodeNics – an array of \*\_nics this cable is connected to. Does not
    include parent
-   VendorName – the name for this port, as seen by the switch
    management software.  Typically this is something along the lines of
    "GigabitEthernet 1/31" ("Berkley-style naming").

  

    "x0c0w0j1" {
        "Parent": "x0c0w0",
        "XName": "x0c0w0",
        "Type": "comptype_mgmt_switch_connector",
        "TypeString": "MgmtSwitchConnector",
        "Class": "river",
        "ExtraProperties": {
            "NodeNics": [
                "x0c0s0b0i0",
                "x0c0s1b0i0",
                ...
            ],
            "VendorName": "GigabitEthernet 1/31"  #Vendor specific port name
        }
    }

  

### \*\_bmc, comptype\_ncard

-   IP6addr – an address, or "DHCPv6".  May be omitted; if so, assumed
    to be DHCPv6
-   IP4addr – an address, or "DHCP".  May be omitted; if so, assumed to
    be DHCP
-   Username – the user name that should be used when accessing the
    device
-   Password – the associated password

  

    "x0c0s0b0" {
        "Parent": "x0c0s0",
        "Children: [
            "x0c0s0b0n0",
            "x0c0s0b0n1",
            ...
        ]
        "XName": "x0c0s0b0",
        "Type": "comptype_ncard",
        "TypeString": "NodeBMC",
        "Class": "mountain",
        "ExtraProperties": {
            "Network": "HMN",
            "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
            "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
            "Username": "user_name",  # to access mgmt interface
            "Password": "vault://tok" # ptr to vault URL to get password
        }
    }

  

### comptype\_bmc\_nic

-   IP6addr – an address, or "DHCPv6".  May be omitted; if so, assumed
    to be DHCPv6
-   IP4addr – an address, or "DHCP".  May be omitted; if so, assumed to
    be DHCP
-   Username - the optional username that should be used when accessing
    the device (or assigned to it)
-   Password - the optional password that should be used when accessing
    the device (or assigned to it)

  

    "x0c0s0b0i0" {
        "Parent": "x0c0s0b0",
        "XName": "x0c0s0b0i0",
        "Type": "comptype_bmc_nic",
        "TypeString": "NodeBMCNic",
        "Class": "river"
        "ExtraProperties": {
            "Network": "HMN",
            "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
            "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
            "Username": "user_name",  # to access mgmt interface
            "Password": "vault://tok" # ptr to vault URL to get password
        }
    }

  

### \*\_nic

-   Networks - a list of networks this NIC is connected to
-   Peers - a list of xnames this nic is directly connected to
-   MACAddr - the MAC address of the NIC (optional)

  

    "x0c0s0b0n0i0" {
        "Parent": "x0c0s0b0n0",
        "XName": "x0c0s0b0n0i0",
        "Type": "comptype_node_nic",
        "TypeString": "NodeNIC",  # ZZZZ case should match others in hmstypes.go
        "Class": "river"
        "ExtraProperties": {
            "MacAddr": "00:11:22:33:44:55",
            "Network": "NMN",
            "Peers": [
                "x0c0r0i0",
                ...
            ]
        }
    }

  

### comptype\_rtrmod

-   PowerConnector - an array of power connectors providing power to
    this device, if any.  Empty for Mountain switch cards

  

    "x0c0r0" {
        "Parent": "x0c0",
        "Children: [
            "x0c0r0j0",
            "x0c0r0j1",
            ...
        ]
        "XName": "x0c0r0",
        "Type": "comptype_rtrmod",
        "TypeString": "RouterModule",
        "Class": "mountain",
        "ExtraProperties": {
            "PowerConnector": [   # list of PDU power connectors we're hooked up to
                "x0m0v0",
                "x0m0v1"
            ]
        }
    }

  

### comptype\_mgmt\_switch, comptype\_compmod

-   PoweredBy - the power connector providing power to this device.
    Required for River hardware, omitted for Mountain hardware.

  

    "x0c0w0" {
        "Parent": "x0c0",
        "Children: [
            "x0c0w0j1",
            "x0c0w0j2",
            ...
        ]
        "XName": "x0c0w0",
        "Type": "comptype_mgmt_switch",
        "TypeString": "MgmtSwitch",
        "Class": "river",
        "ExtraProperties": {
            "Network": "HMN",
            "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
            "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
            "Username": "user_name",  # to access mgmt interface
            "Password": "vault://tok" # ptr to vault URL to get password
            "PowerConnector": "x0m0v1"
        }
    },

    "x0c0s0" { 
        "Parent": "x0c0",
        "XName": "x0c0s0",
        "Type": "comptype_compmod",
        "TypeString": "ComputeModule",
        "Class": "river"
        "ExtraProperties": {
            "PoweredBy": "x0m0v0" # valid for river only
        }
    }

  

### comptype\_node

-   NID - the numeric NID assigned to this node. Should not be specified
    unless role is "compute", when it is required.
-   Role - the role type assigned to this node,
    from <https://stash.us.cray.com/projects/HMS/repos/hms-common/browse/pkg/base/hmstypes.go#752>

  

    "x0c0s0b0n0" {
        "Parent": "x0c0s0b0",
        "XName": "x0c0s0b0n0",
        "Type": "comptype_node",
        "TypeString": "Node",
        "Class": "river"
        "ExtraProperties": {
            "NID": 1234
            "Role": "Compute"
        }
    }

  

## The Network Object

In addition to tracking hardware, SLS must also track the network that
items are connected to.  This means it also has to track the networks
available in the system.  For version 1.0, this is represented by a JSON
object with the following properties:

-   Name – a short name for the network.  This is the key that will be
    used to identify this network elsewhere.  Must be unique, should not
    contain spaces.  eg: "HMN".
-   FullName – The name of the network in a human readable form, eg
    "Hardware Management Network"
-   IPRanges – an array of strings in CIDR notation marking which IP
    addresses are assigned to this network.
-   Type – What type of hardware makes up the network.  Valid values are
    "slingshot10", <span class="inline-comment-marker"
    ref="73e4f840-f503-420d-b21f-f78707c7d70d">"cassini"</span>,
    "ethernet", "OPA", "infiniband", "mixed"

  

    "HSN" {
        "Name": "HSN",                            # Short name, arbitrary     
        "FullName": "Rosetta High Speed Network", # Descriptive name, arbitrary
        "IPRanges": [          # CIDRs
            "10.1.1.0/24",
            "10.1.2.0/24",
            ...
        ],
        "Type": "Slingshot10"  # slingshot10,slingshot11,ethernet,OPA,infiniband,mixed
    }

  

# RESTful Interface

## /<span class="inline-comment-marker" ref="8eb1f182-0595-4602-b1ad-9da0f568649e">ready</span>

### GET

GET returns a JSON array with the following keys:

-   Ready: either true or false.  True indicates SLS is ready to go
-   Reason: If ready is false, contains a human-readable reason SLS is
    not ready to go.  If ready is true, empty or omitted.
-   Code: If ready is false, contains a machine-readable code why SLS is
    not ready to go.  If ready is true, empty or omitted.

## <span class="inline-comment-marker" ref="4bea4285-abe2-402d-ba88-6d28c66fbd6b">/version</span>

### GET

Contains information on the current version of this mapping.  GET-only. 
Information returned is a JSON array with two keys:

-   Counter: a monotonically increasing counter.  This counter is
    incremented every time a change is made to the map stored in SLS. 
    This shall be 0 if no data is uploaded to SLS
-   LastUpdated: an ISO 8601 datetime representing the time of the last
    change to SLS.  This shall be set to the Unix Epoch if no data has
    ever been stored in SLS.

## /hardware

### POST

Create a new hardware entry.  In the future, parent entries may be
created, but this is not going to be implemented in SLS1.0

If this is one of a small set of hardware types supported by HSM (eg:
comptype\_node), SLS shall create the item in HSM.  See "Interaction
with HSM" below.

## /hardware/{xname}

### GET

Return information for the hardware with xname `xname`.  All properties
should be returned as a JSON array.

### <span class="inline-comment-marker" ref="5b825670-8ce8-4ff0-8c60-3a3eb66379a3">PATCH</span>

Update information for xname.  The properties xname, type, human\_type,
children, and parent are immutable; attempting to alter these will
result in an error.

### DELETE

Remove xname from the system.

## /networks

### GET

Returns a JSON list of the networks available in the system.  Return
value is an array of strings with each string representing the name
field of the network object

### POST

Allows the creation of new networks.  Must include all fields at the
time of upload.

## /networks/{name}

### GET

Gets the network object for this network

### PATCH

Allows updates to the named network object

### DELETE

Deletes the associated network

## /search

### GET

Uses HTTP query parameters to find hardware entries with matching
properties.  Returns a JSON list of xnames. If multiple query parameters
are passed, any returned hardware must match all parameters.

For example, a <span class="inline-comment-marker"
ref="526ee414-2b45-494b-954b-60572c414694">query string</span> of
"?parent=x0" would return a list of all children of cabinet x0.  A query
string of "?parent=x0&type=comptype\_node" would return a list of all
compute nodes in cabinet x0.

Valid query parameters are: xname, parent, class, type,
power\_connector, node\_nics, networks, peers

## /dumpstate

### GET

Dumps the current database state of the service. Should be an etcd
database export that can be re-imported or another key-value format we
could re-import.

## /<span class="inline-comment-marker" ref="b2252018-b661-48e2-a182-6c1867bfbbaa">loadstate</span>

### POST

Overwrite the current database with the contents of the posted data. 
The posted data should be a state dump from /dumpstate.

# Initialization

A tool will be built to initialize SLS based on an input file.  This
will likely be a site.yaml-based tool, though details are TBD.

# Interaction with HSM

HSM will periodically poll for changes to the SLS data and fetch any new
information it needs.  

# Upload Tool

In it's final packaging, SLS shall have the ability to upload a database
via /loadstate.  However, something needs to generate the state to load
and actually perform the upload.  These tools are the <span
class="inline-comment-marker"
ref="cc270727-1c82-4595-ac84-8ce042ce9532">SLSGenerator</span> and
SLSUploader, respectively.

## SLSUploader

Let's start with the simpler of the two: the SLS uploader.  Under our
current Kubernetes regime, this is best built as an init job that
performs several tasks:

1.  Checks if data has already been uploaded to SLS.  It can do so by
    checking /version.  If so, the rest of this is skipped.
2.  If no data has yet been uploaded, it reads the contents of a
    configmap and uploads those.

Point 2 implies we need an agreed format for the SLSUploader to read the
configmap.  Further, our lives will be simpler if this is the same
format as /loadstate and can be passed through.  This would allow
SLSUploader to be a simple bash script...

## SLS Upload Format

In order to upload data to SLS, it needs to be in a consistent format. 
Therefore, the following format is offered as a suggestion:

    {
        "Hardware": {
            # HSN connector, River or Mountain
            "x0c0r0j0": {
                "Parent": "x0c0r0",
                "XName": "x0c0r0j0",
                "Type": "comptype_hsn_connector",
                "TypeString": "HSNConnector",
                "Class": "river",
                "ExtraProperties": {
                    "NodeNics": [        #Node NICs this HSN port connects to
                        "x0c0s0b0n0h0",
                        "x0c0s0b0n1h1",
                        ...
                    ]
                }
            },
        
            # MGMT switch
            "x0c0w0": {
                "Parent": "x0c0",
                "Children: [
                    "x0c0w0j1",
                    "x0c0w0j2",
                    ...
                ]
                "XName": "x0c0w0",
                "Type": "comptype_mgmt_switch",
                "TypeString": "MgmtSwitch",
                "Class": "river",
                "ExtraProperties": {
                    "Network": "HMN",         # or NMN
                    "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
                    "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
                    "Username": "user_name",  # to access mgmt interface
                    "Password": "vault://tok" # ptr to vault URL to get password
                }
            },
        
            # MGMT switch connector
            "x0c0w0j1": {
                "Parent": "x0c0w0",
                "XName": "x0c0w0",
                "Type": "comptype_mgmt_switch_connector",
                "TypeString": "MgmtSwitchConnector",
                "Class": "river",
                "ExtraProperties": {
                    "NodeNics": [
                        "x0c0s0b0i0",
                        "x0c0s1b0i0",
                        ...
                    ],
                    "VendorName": "GigabitEthernet 1/31"  #Vendor specific port name
                }
            },
        
            # Mountain cabinet (can be used for River rack as well without ExtraProperties)
            "x1000": {
                "Parent": "s0",
                "Children: [
                    "x1000c0",
                    "x1000c1",
                    ...
                ],
                "XName":"x1000",
                "Type": "comptype_cabinet",
                "TypeString": "Cabinet",
                "Class": "mountain",
                "ExtraProperties": {
                    "Network": "HMN",
                    "IP6Prefix": "fd66:0:0:0",
                    "IP4Base": "10.1.2.100/26",
                    "MACprefix": "02"
                }
            }

            # River BMC or Mountain controller cards
            "x0c0s0b0": {
                "Parent": "x0c0s0",
                "Children: [
                    "x0c0s0b0n0",
                    "x0c0s0b0n1",
                    ...
                ]
                "XName": "x0c0s0b0",
                "Type": "comptype_ncard",
                "TypeString": "NodeBMC",
                "Class": "mountain",
                "ExtraProperties": {
                    "Network": "HMN",
                    "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
                    "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
                    "Username": "user_name",  # to access mgmt interface
                    "Password": "vault://tok" # ptr to vault URL to get password
                }
            },
        
            "x0c0r0b0": {
                "Parent": "x0c0r0",
                "XName": "x0c0r0b0",
                "Type": "comptype_rtr_bmc",
                "TypeString": "RouterBMC",
                "Class": "river",
                "ExtraProperties": {
                    "Network": "HMN",
                    "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
                    "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
                    "Username": "user_name",  # to access mgmt interface
                    "Password": "vault://tok" # ptr to vault URL to get password
                }
            },
        
            "x0c0b0": {
                "Parent": "x0c0",
                "Children: [
                    "x0c0s0",
                    "x0c0s1",
                    "x0c0r0",
                    "x0c0r1",
                    ...
                ],
                "XName": "x0c0b0",
                "Type": "comptype_chassis_bmc",
                "TypeString": "ChassisBMC",
                "Class": "mountain",
                "ExtraProperties": {
                    "Network": "HMN",
                    "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
                    "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
                    "Username": "user_name",  # to access mgmt interface
                    "Password": "vault://tok" # ptr to vault URL to get password
                }
            },
        
            # River BMC/Mountain controller NICs (HW Mgmt Network)
            "x0c0s0b0i0": {
                "Parent": "x0c0s0b0",
                "XName": "x0c0s0b0i0",
                "Type": "comptype_bmc_nic",
                "TypeString": "NodeBMCNic",
                "Class": "river",
                "ExtraProperties": {
                    "Network": "HMN",
                    "MACAddr: "00:11:22:33:44:55",
                    "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
                    "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
                    "Username": "user_name",  # to access mgmt interface
                    "Password": "vault://tok" # ptr to vault URL to get password
                }
            },
        
            "x0c0r0b0i0": {
                "Parent": "x0c0r0b0",
                "XName": "x0c0r0b0i0",
                "Type": "comptype_rtr_bmc_nic",
                "TypeString": "RouterBMCNic",
                "Class": "mountain",
                "ExtraProperties": {
                    "Network": "HMN",
                    "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
                    "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
                    "Username": "user_name",  # to access mgmt interface
                    "Password": "vault://tok" # ptr to vault URL to get password
                }
            },
        
            "x0c0b0i0": {
                "Parent": "x0c0b0",
                "XName": "x0c0b0i0",
                "Type": "comptype_chassis_bmc_nic",
                "TypeString": "Chassis,
                "ExtraProperties": {
                    "Network": "HMN",
                    "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
                    "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
                    "Username": "user_name",  # to access mgmt interface
                    "Password": "vault://tok" # ptr to vault URL to get password
                }
            },
        
            # Node NICs (Node Mgmt Network), river and mountain nodes
            "x0c0s0b0n0i0": {
                "Parent": "x0c0s0b0n0",
                "XName": "x0c0s0b0n0i0",
                "Type": "comptype_node_nic",
                "TypeString": "NodeNIC",
                "Class": "river",
                "ExtraProperties": {
                    "MACAddr: "00:11:22:33:44:55",
                    "Network": "NMN",
                    "Peers": [
                        "x0c0r0i0",
                        ...
                    ]
                }
            }
        
            # Node HSN NICs (HSN Network), river and mountain nodes
            "x0c0s0b0n0h0": {
                "Parent": "x0c0s0b0n0",
                "XName": "x0c0s0b0n0h0",
                "Type": "comptype_node_hsn_nic",
                "TypeString": "NodeHsnNic",
                "Class": "mountain",
                "ExtraProperties": {
                    "Network": "HSN",
                    "MACAddr: "00:11:22:33:44:55",
                    "Peers": [       # HSN endpoints we're directly connected to
                        "x0c0r0i10",
                        ...
                    ]
                }
            }
        
            # HSN Switches (Mountain switch blade or river HSN TOR)
            "x0c0r0": {
                "Parent": "x0c0",
                "Children: [
                    "x0c0r0j0",
                    "x0c0r0j1",
                    ...
                ]
                "XName": "x0c0r0",
                "Type": "comptype_rtrmod",
                "TypeString": "RouterModule",
                "Class": "mountain",
                "ExtraProperties": {
                    "PowerConnector": [   # list of PDU power connectors we're hooked up to
                        "x0m0v0",
                        "x0m0v1"
                    ]
                }
            }
        
            # Mountain or River node blade, including non-SMS NCN
            "x0c0s0": {
                "Parent": "x0c0",
                "XName": "x0c0s0",
                "Type": "comptype_compmod",
                "TypeString": "ComputeModule",
                "Class": "river"
                "ExtraProperties": {
                    "PoweredBy": "x0m0v0"
                }
            },
        
            # Mountain or River Nodes, including non-SMS NCN
            "x0c0s0b0n0": {
                "Parent": "x0c0s0b0",
                "XName": "x0c0s0b0n0",
                "Type": "comptype_node",
                "TypeString": "Node",
                "Class": "river",
                "ExtraProperties": {
                    "NID": 1234,
                    "Role": "Compute"
                }
            }
        
            # River Power Distribution Unit
            "x0m0p0": {
                "Parent": "x0m0",
                "Children: [
                    "x0m0p0v0",
                    "x0m0p0v1",
                    ...
                ]
                "XName": "x0m0p0",
                "Type": "comptype_cab_pdu",
                "TypeString": "CabinetPDU",
                "Class": "river"
            },
        
            # PDU NIC?
            "x0m0p0i0": {
                "Parent": "x0m0p0",
                "XName": "x0m0p0i0",
                "Type": "comptype_cab_pdu_nic",
                "TypeString": "CabinetPDUNic",
                "Class": "river",
                "ExtraProperties": {
                    "Network": "HMN",
                    "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
                    "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
                    "Username": "user_name",  # to access mgmt interface
                    "Password": "vault://tok" # ptr to vault URL to get password
                }
            },
        
            # Mountain CDU mgmt switch
            "d0w0": {
                "Parent": "d0",
                "Children: [
                    "d0w0j0",
                    "d0w0j1",
                    ...
                ]
                "XName": "d0w0",
                "Type": "comptype_cdu_mgmt_switch",
                "TypeString": "CDUMgmtSwitch",
                "Class": "mountain",
                "ExtraProperties": {
                    "Network": "HMN",
                    "IP6addr": "DHCPv6",      # or IPV6 addr, may be omitted, default to DHCPv6
                    "IP4addr": "10.1.1.1",    # or IPV4 addr, may be ommited, default to DHCPv4
                    "Username": "user_name",  # to access mgmt interface
                    "Password": "vault://tok" # ptr to vault URL to get password
                }
            },
        
            # River PDU power connector
            "x0m0p0v0": {
                "Parent": "x0m0p0",
                "XName": "x0m0p0v0",
                "Type": "comptype_cab_pdu_pwr_connector",
                "TypeString": "CabinetPDUPowerConnector",
                "Class": "river",
                "ExtraProperties": {
                    "PoweredBy": "x0c0s0v0"   # what this power connector connects to
                }
            },
        
            # River node/enclosure power connector
            "x0c0s0v0": {
                "Parent": "x0c0s0",
                "XName": "x0c0s0v0",
                "Type": "comptype_compmod_power_connector",
                "TypeString": "NodePowerConnector",
                "Class": "river",
                "ExtraProperties": {
                    "PoweredBy": "x0m0p0v2"  # what PDU socket this power connector is connected to
                }
            },
        
            # River TOR power connector
            "x0c0r0v0": {
                "Parent": "x0c0r0",
                "XName": "x0c0r0v0",
                "Type": "comptype_rtrmod_power_connector",
                "TypeString": "RouterPowerConnector",
                "Class": "river",
                "ExtraProperties": {
                    "PoweredBy": "x0m0p0v1"  # what PDU socket this power connector is connected to
                }
            },
        },
        "Networks": {
            # Network Descriptions
            "HSN": {
                "Name": "HSN",                            # Short name, arbitrary
                "FullName": "Rosetta High Speed Network", # Descriptive name, arbitrary
                "IPRanges": [          # CIDRs
                    "10.1.1.0/24",
                    "10.1.2.0/24",
                    ...
                ],
                "Type": "Slingshot10"  # slingshot10,slingshot11,ethernet,OPA,infiniband,mixed
            },
        
            "HMN": {
                "Name": "HMN",                             # Short name, arbitrary
                "FullName": "Hardware Management Network", # Descriptive name, arbitrary
                "IPRanges": [          # CIDRs
                    "10.100.1.0/26",
                    "10.100.2.0/26",
                    ...
                ],
                "Type": "Ethernet"     # slingshot10,slingshot11,ethernet,OPA,infiniband,mixed
            },
        ...
        }
    }

  

That is, a JSON object, where each property's key is the xname of the
item contained therein.  The JSON representation of the item is another
object with attributes and values equal to the attributes defined above

## Attachments:

<img src="images/icons/bullet_blue.gif" width="8" height="8" />
[sls-objects-full](attachments/148473555/148475763)
(application/gliffy+json)  
<img src="images/icons/bullet_blue.gif" width="8" height="8" />
[sls-objects-full.png](attachments/148473555/148475764.png)
(image/png)  
<img src="images/icons/bullet_blue.gif" width="8" height="8" />
[sls-objects-full](attachments/148473555/148475761)
(application/gliffy+json)  
<img src="images/icons/bullet_blue.gif" width="8" height="8" />
[sls-objects-full.png](attachments/148473555/148475762.png)
(image/png)  

## Comments:

<table data-border="0" width="100%">
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><span id="comment-149883819"></span>
<p>Should this document talk about how SLS get's it's items for the system? (i.e yaml file)</p>
<p>Should this document talk about how it will update HSM with said information?</p>
<p><br />
</p>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by rfrost at Aug 29, 2019 08:22
</div></td>
</tr>
<tr class="even">
<td style="border-top: 1px dashed #666666"><span id="comment-149884613"></span>
<p>Concur, this should all be in this doc.</p>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by mpkelly at Aug 30, 2019 10:34
</div></td>
</tr>
<tr class="odd">
<td style="border-top: 1px dashed #666666"><span id="comment-149906872"></span>
<p>The Shasta Tools v1.0.16 already produce a YAML file containing each &lt;comptype_node_hsn_nic&gt; in a Slingshot network. A sample from Coke is:</p>
<div class="code panel pdl" style="border-width: 1px;">
<div class="codeContent panelContent pdl">
<pre class="syntaxhighlighter-pre" data-syntaxhighlighter-params="brush: java; gutter: false; theme: Confluence" data-theme="Confluence"><code>- x3000c0s9b3n3h0:

    comptype_hsn_connector: x3000c0r19j16

    comptype_node: x3000c0s9b3n3

    group_id: 0

    hsn_ip_addr: 169.0.0.96

    hsn_mac_addr: 02:00:00:00:00:5f

    interface_cfg_name: hsn0

    port_designator: PF15B

    port_number: 31

    switch_number: 1

- x3000c0s9b4n4h0:

    comptype_hsn_connector: x3000c0r18j14

    comptype_node: x3000c0s9b4n4

    group_id: 0

    hsn_ip_addr: 169.0.0.13

    hsn_mac_addr: 02:00:00:00:00:0c

    interface_cfg_name: hsn0

    port_designator: PF06A

    port_number: 12

    switch_number: 0</code></pre>
</div>
</div>
<p><br />
</p>
<p>This can file output be generated for any system and used immediately.</p>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by mvera at Oct 02, 2019 16:14
</div></td>
</tr>
<tr class="even">
<td style="border-top: 1px dashed #666666"><span id="comment-149906878"></span>
<p>The Shasta Tools v1.0.16 also generate a JSON file with differing field names for the same data. A similar example from Coke:</p>
<div class="code panel pdl" style="border-width: 1px;">
<div class="codeContent panelContent pdl">
<pre class="syntaxhighlighter-pre" data-syntaxhighlighter-params="brush: java; gutter: false; theme: Confluence" data-theme="Confluence"><code>    {

        &quot;connector&quot;: &quot;x3000c0r22j16&quot;, 

        &quot;hsn_ip_addr&quot;: &quot;169.0.16.31&quot;, 

        &quot;grpId&quot;: 2, 

        &quot;portNum&quot;: 30, 

        &quot;swcNum&quot;: 0, 

        &quot;interface_cfg_name&quot;: &quot;hsn0&quot;, 

        &quot;hsn_mac_addr&quot;: &quot;02:00:00:00:10:1e&quot;, 

        &quot;pf&quot;: &quot;PF15A&quot;, 

        &quot;xname&quot;: &quot;x3000c0s13b1n1h0&quot;

    }, 

    {

        &quot;connector&quot;: &quot;x3000c0r22j16&quot;, 

        &quot;hsn_ip_addr&quot;: &quot;169.0.16.32&quot;, 

        &quot;grpId&quot;: 2, 

        &quot;portNum&quot;: 31, 

        &quot;swcNum&quot;: 0, 

        &quot;interface_cfg_name&quot;: &quot;hsn0&quot;, 

        &quot;hsn_mac_addr&quot;: &quot;02:00:00:00:10:1f&quot;, 

        &quot;pf&quot;: &quot;PF15B&quot;, 

        &quot;xname&quot;: &quot;x3000c0s13b3n3h0&quot;

    }</code></pre>
</div>
</div>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by mvera at Oct 02, 2019 16:16
</div></td>
</tr>
<tr class="odd">
<td style="border-top: 1px dashed #666666"><span id="comment-154536776"></span>
<p><a href="https://connect.us.cray.com/confluence/display/~mvera" class="confluence-userlink user-mention">Michael Vera</a>: I've just updated the upload file format slightly.  I did two things:</p>
<ol>
<li>Created separate sections for "Hardware" and "Networks".  Mingling the two didn't really make sense.</li>
<li>Fixed a syntax error re: keys for top-level objects</li>
</ol>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by spresser at Oct 10, 2019 08:14
</div></td>
</tr>
<tr class="even">
<td style="border-top: 1px dashed #666666"><span id="comment-186422634"></span>
<p>What Shasta Tools are referred to here?</p>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by cbloxham at Sep 08, 2020 10:39
</div></td>
</tr>
</tbody>
</table>

Document generated by Confluence on Jan 14, 2022 07:17

[Atlassian](http://www.atlassian.com/)
