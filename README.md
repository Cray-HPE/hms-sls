# System Layout Service (SLS)

The System Layout Service (SLS) holds information about a Cray Shasta system "as configured".  That is, it contains information
about how the system was designed.  SLS reflects the system if all hardware were present and working.  SLS does not keep track
of hardware state or identifiers; for that type of information, see Hardware State Manager (HSM).

## How it Works

SLS is mostly an API to allow access to a database.  Users make queries on the basis of xname and are given back information
about the system.  Each xname is matched with corresponding parent and child information, allowing the simple traversal of the
entirety of the system.

## Configuration

TBD.  Initial configuration will be uploaded by an init container.

## SLS CT Testing

In addition to the service itself, this repository builds and publishes cray-sls-test images containing tests that verify SLS
on live Shasta systems. The tests are invoked via helm test as part of the Continuous Test (CT) framework during CSM installs
and upgrades. The version of the cray-sls-test image (vX.Y.Z) should match the version of the cray-sls image being tested, both
of which are specified in the helm chart for the service.

## More information

* [SLS design documentation](docs/SLS-design-doc.md)
* [SLS Overview](docs/overview.md)

## Future Work

TBD.