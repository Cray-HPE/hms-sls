# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [2.11.0] - 2025-06-18

### Added

- IPv6 support in SLS subnets

### Changed

- Deprecated `IPV4Subnet` in favor of `IPSubnet` (deprecated alias `IPV4Subnet` points to `IPSubnet`)

## [2.10.0] - 2025-06-09

### Changed

- Changed tests to allow the fields: CIDR6, Gateway6, and IPAddress6.
- Added the new fields to the API docs

## [2.9.0] - 2025-05-02

### Update

- Updated module and image dependencies to latest versions
- Update version of Go to v1.24
- Explicitly drain and close all http request and response bodies
- Internal tracking ticket: CASMHMS-6399

## [2.8.0] - 2025-02-18

### Security

- Update image and module dependencies
- Various code changes to accomodate module updates
- Added image-pprof Makefile support
- Resolved build warnings in Dockerfiles and docker compose files
- Switched CT tests from using RTS to RIE

## [2.7.0] - 2025-01-08

### Added

- Added support for ppprof builds

## [2.6.0] - 2024-12-03

### Changed

- Updated go to 1.23
- Updated hms-s3, hms-base, and hms-xname

## [2.5.0] - 2024-08-23

### Changed
- Changed the tests to allow IPv6 CIDR addresses

## [2.4.0] - 2023-09-25

### Added
- Updated the ct tests to support the VirtualNode type

## [2.3.0] - 2023-08-30

### Added
- Added support for VirtualNode Hardware entries

## [2.2.0] - 2023-06-23

### Added
- Created a simple benchmark tool that can stress test SLS with many concurrent requests. 

### Changed
- Updated Golang to 1.19
- Introduced a caching HTTP middleware. This feature is disabled by default.
- Database interaction improvements
  - Pass contexts from the HTTP layer into database layer to cancel database operations if the HTTP requests is canceled.
  - Database functions now accept `*tx.SQL` to allow use within database transactions.  
  - Created `*Context` functions that warp the database functions within a transaction that is canceled by the provided context.
  - `SetGenericHardwareContext` and `SetNetworkContext` were added to atomicity create or update hardware and network information, instead of running 2-3 database queriers that could leave the database in an inconsistent state if canceled.  
  - `SearchGenericHardware` was modified to perform a single query like `GetAllGenericHardware`, as it was running additional SQL queries while processing data from another query. This type of database interaction can cause SLS to become deadlocked. 
  - Removed `getChildrenForXname` as it is no longer required.
- Datastore changes
  - Pass in contexts from HTTP layer for use with the database layer. 
  - Removed xname parameter from `SetXname` as it was unused, as the xname was being provided from the generic hardware object.
- SLS Client
  - Added `GetNetworks` method to call `/v1/networks`
  - Added `GetNetwork` method to call `/v1/networks/$NETWORK_NAME`

## [2.1.0] - 2023-01-26

### Fixed

- Language linting of API spec file (no content changes)

## [2.0.0] - 2022-11-30

### Changed
- Removed encrypted dumpstate and loadstate APIs

## [1.29.0] - 2022-09-30

### Changed
- added many API tests using Tavern.

## [1.28.0] - 2022-09-12

### Changed
- CASMHMS-5695 - Improved the performance getting the hardware components

## [1.27.0] - 2022-08-26

### Changed
- CASMHMS-5696 - Disallow Networks with empty names

## [1.26.0] - 2022-08-25

### Changed
- CASMHMS-4267 - Changed loadstate to validate

## [1.25.0] - 2022-08-24

### Changed
- CASMHMS-5691: Added the slingshot11 network type

## [1.24.0] - 2022-08-09

### Changed
- CASMINST-3902: Expanded the SLS client to perform dumpstate and PUT to networks.

## [1.23.0] - 2022-07-27

### Changed
- CASMINST-3788: Add SLS Migrator to deal with malformed data from older CSM releases. 
  - When malformed liquid-cooled Chassis data is encountered a corresponding ChassisBMC will be created.
  - When malformed xname derivived fields (`Parent`, `Type`, `TypeString`) are encountered the object will be PUT back into SLS to recalculate the fields.
- Added CDUMgmtSwitch as an acceptable type to the SLS CT functional tests.

## [1.22.0] - 2022-07-18

### Changed
- CASMHMS-5488: Changed the way the SQL is built for the search API

## [1.21.0] - 2022-06-24

### Changed
- CASMHMS-5291: Add Model field to the `ComptypeCabinet` structure.
- Updated the hms-xname package to 1.1.0.
- Updated references of hms-base to the v2 version of the package. 

## [1.20.0] - 2022-06-23

### Changed

- Updated CT tests to hms-test:3.1.0 image as part of Helm test coordination.

## [1.19.0] - 2022-05-19

### Changed

- Updated SLS to build using GitHub Actions instead of Jenkins.
- Pull images from artifactory.algol60.net instead of arti.dev.cray.com.
- Added a runCT.sh script that can run the CT tests in a docker-compose environment.

## [1.18.0] - 2022-04-11

### Changed

- CASMHMS-5350 - Improved swagger documentation

## [1.17.0] - 2022-03-23

### Changed

- CASMHMS-5259 - Added validation for IPRanges when setting networks

## [1.16.0] - 2022-02-03

### Changed

- CASMHMS-4670 - PUT and POST /hardware validation and http response improvements.

## [1.15.0] - 2022-01-31

### Changed

- CASMHMS-4671 - Hardware Search API now returns 400 for bad requests instead of the 500 HTTP status.

## [1.14.0] - 2022-01-25

### Added

- CASMHMS-4270 - Support Cabinet and CDU hardware objects
- CASMHMS-4669 - Support HL Switch and RTR TOR FPGA hardware objects.

## [1.13.0] - 2021-12-17

### Added

- CASMINST-3617 - Added PeerASN and MyASN to NetworkExtraProperties struct
- CASMNET-697 - Added MetalLBPoolName to IPV4Subnet struct

## [1.12.0] - 2021-10-27

### Added

- CASMNET-692 - Added Bifurcated CAN default route toggle.

## [1.11.0] - 2021-10-27

### Added

- CASMHMS-5055 - Added SLS CT test RPM.

## [1.10.6] - 2021-09-16

### Changed

- Changed the docker image to run as the user nobody

## [1.10.5] - 2021-08-10

### Changed

- Added GitHub configuration files.

## [1.10.4] - 2021-07-26

### Changed

- Replaced all old stash paths with github.com

## [1.10.3] - 2021-07-19

### Changed

- Add support for building within the CSM Jenkins.

## [1.10.2] - 2021-07-02

### Security

- CASMHMS-4929 - Enable automatic postgres backups for SLS.

## [1.10.1] - 2021-06-30

### Security

- CASMHMS-4898 - Updated base container images for security updates.

## [1.10.0] - 2021-06-18

### Changed
- Bump minor version for CSM 1.2 release branch

## [1.9.0] - 2021-06-18

### Changed
- Bump minor version for CSM 1.1 release branch

## [1.8.10] - 2021-05-13

### Changed
- Added aliases to ComptypeNodeBmc and added new struct for ComptypeChassisBmc to support aliases for that as well if necessary.

## [1.8.9] - 2021-05-05

### Changed
- Updated docker-compose files to pull images from Artifactory instead of DTR.

## [1.8.8] - 2021-05-04

### Changed
- CASMINST-2121: Added new fields to the IPV4Subnet struct to support
  uai_macvlan in csi

## [1.8.7] - 2021-04-27

### Changed
- CASMHMS-4765: Set a limit for the maximum number of database connections SLS can have open.

## [1.8.6] - 2021-04-20

### Changed
- Updated Dockerfile to pull base images from Artifactory instead of DTR.

## [1.8.5] - 2021-04-06

### Changed
- CASMHMS-4600 - Fixed an issue where the Hardware search API did not accept `comptype_hl_switch` and `comptype_cdu_mgmt_switch` as valid values to the `type` query param.
- CASMHMS-4578/CASMHMS-4749 - Update the cray-service chart to 2.4.7 to address postgres security vulnerabilities and wait-for-postgres resource limit changes..
- Fixed an issue where SLS did not have `comptype_cab_pdu_pwr_connector` properly defined.

## [1.8.4] - 2021-03-31

### Changed
- CASMHMS-4605 - Update the loftsman/docker-kubectl image to use a production version.

## [1.8.3] - 2021-03-19

### Changed
- CASMHMS-4554: Scale SLS to 3 replicas with anti-affinity to prevent multiple SLS pods running on the same worker node.

## [1.8.2] - 2021-03-09

### Changed
- CASMINST-1546: Improved error handling in the SLS loader job. Modified the process of determining the IP address of rgw-vip.nmn to be more robust.

## [1.8.1] - 2021-02-03

### Added
- Adding the runSnyk.sh script which was missed previosuly.

## [1.8.0] - 2021-02-01

### Changed
- Update License/Copyright info, re-vendor go packages.

## [1.7.1] - 2021-01-26

### Changed
- CASMINST-1126: Pickup the latest cray-service base chart to pick the wait-for-postgres jobs to prevent these jobs from getting OOMKilled

## [1.7.0] - 2021-01-14

### Changed

- Updated license file.


## [1.6.0] - 2021-01-08
### Changed
- CASMINST-759: Use the livecd nameserver to determine the IP address of the S3 endpoint. In order for DNS name resolution in k8s to work properly SLS needs to be populated with data, so that Ubound manager job can setup DNS records. However, when the SLS loader job first runs unbound is empty and is unable to resolve the S3 endpoint.

## [1.5.4] - 2020-12-02
### Changed
- CASMHMS-4266 - Added support to SLS for MgmtHLSwitch & CDUMgmtSwitch, and updated HMS Base to v1.8.4.

## [1.5.3] - 2020-10-29

### Security

- CASMHMS-4148 - Update go module vendor code for security fix.

## [1.5.2] - 2020-10-27

### Changed
- CASMHMS-4055 - The SLS Loader job will now only upload the SLS input file once. The new default behavior of the SLS loader job is to upload the SLS file if the SLS S3 bucket does not contain the special file `uploaded`. If that file is not present in S3 then the SLS loader will load the SLS file into SLS, otherwise the loader will perform a no-op. If the that file is present in S3, then the loader job will do nothing. After the loader performs the SLS loadstate, it will create the `uploaded` file in the SLS S3 bucket.
- CASMHMS-4163: Update cray-service-base char to the latest 2.2.0 version.

## [1.5.1] - 2020-10-22

### Security

- CASMHMS-4105 - Updated base Golang Alpine image to resolve libcrypto vulnerability.

## [1.5.0] - 2020-10-19

### Changed
- CASMHMS-4099 - The SLS Network structures have been greatly enriched. The base Network structure has not been modified, and all new networking information has been added to the network's extra properties. Networks are now meant to represent a IPv4 Network, and each IPv4 network can describe the IPv4 subnets within the network. IP reservations can also be described within a IPv4 subnet.
- CASMHMS-4100 - Download the pre-generated `sls_input_file.json` from the SLS S3 bucket.  SLS no longer generates the SLS input file within the SLS Init/Load job, instead the SLS file is generated off of the system and then uploaded into the SLS S3 bucket. This is the new behavior in Shasta v1.4 and forward.

## [1.4.1] - 2020-10-21

### Changed
- Upgraded the cray-service chart to the latest version

## [1.4.0] - 2020-09-15

### Changed
- CASMCLOUD-1023 - Updated cray-service base chart to the latest 2.x version. This new version of the cray service chart now supports Helm v3.
- Modified containers/init containers, volume, and persistent volume claim value definitions to be objects instead of arrays

## [1.3.7] - 2020-09-10

### Security

-  CASMHMS-3996 - Updated hms-sls to use trusted baseOS images.

## [1.3.6]

### Fixed

- CASMHMS-3985 - fixed switch xname generation to use destination rack rather than source.

## [1.3.5] - 2020-08-19

### Fixed

- CASMHMS-3792 - Improved support for PDUs. Fixed Management switch connectors for PDUs to use the correct xname for the PDU.

## [1.3.4]

### Changed

- CASMHMS-3914 - moved CMC BMC number to 999.

## [1.3.3]

### Fixed

- CASMHMS-3768 - made parsing tolerate unknown hardware better.

## [1.3.2] - 2020-07-16

### Fixed

- CASMHMS-3768 - Fixed bug where the config parser would fall right over when it got to something not in a U.

## [1.3.1] - 2020-06-30

### Fixed

- CASMHMS-3674 - fixed parsing bug.

## [1.3.0] - 2020-06-29

### Added

- CASMHMS-3611 - Added CT smoke test for SLS.

## [1.2.3]

### Fixed

- CASMHMS-3648 - fixed processing of xnames for Columbia switches.

## [1.2.2]

### Fixed

- CASMHMS-3639 - fixed config parser bug.

## [1.2.1]

### Fixed

- CASMHMS-3635 - made file getting from S3 try forever.

## [1.2.0]

### Added

- CASMHMS-3550 - added all logic relating to downloading files from S3, generating SLS config, and pushing that into SLS.

## [1.1.0]

### Added

- CASMHMS-3466 - added a lot of the parsing logic for the new config files.

## [1.0.1]

### Fixed

- CASMHMS-3526 - fixed job cleanup.

## [1.0.0]

### Changed

- CASMHMS-3456 - added ExtraProperties section to Networks object.
- Changed migration logic slightly to requested version instead of up all the time.
- Updated to version 1.5.0 of the base `cray-service` chart.

## [0.9.5]

### Changed

- CASMHMS-3263 - updated cray-service base chart to enable online install upgrade/rollback

## [0.9.4]

### Changed

- CASMHMS-2965 - use golang based image for build-base to align with other services.

## [0.9.3]

### Changed

- CASMHMS-2965 - use trusted baseOS image.

## [0.9.2]

### Changed

- Moved the SLS loader out of the ansible installer and into the SLS helm chart.
- Removed the use of the wait-for-postgres job, which will be removed from the base chart.

## [0.9.1]

### Changed

- CASMHMS-2900 - Updated swagger file to fix openapi conversion issues and include missed commands.

## [0.9.0]

### Added

- Added SLS loader.

## [0.8.2]

### Changed

- Made SLS tolerate not having keys for load/dump state.

## [0.8.1]

### Changed

- Changed the Postgres configuration during unit tests to allow connections without passwords. A breaking change was made to the official postgres image to require a password by default.

## [0.8.0]

### Changed

- CASMHMS-2641 - added liveness, readiness, and health endpoints.

## [0.7.1]

### Changed

- Updated hms-common lib.

## [0.7.0]

### Added

- Added encrypted dump/load of SLS. This can be used by:
    1. Generate a public/private key pair:

       ```openssl rsa -in private.pem -outform PEM -pubout -out public.pem```
    2. Dump encrypted:

        ```
        curl -X POST \
          http://localhost:8376/v1/dumpstate \
          -H 'Accept: */*' \
          -H 'Accept-Encoding: gzip, deflate' \
          -H 'Cache-Control: no-cache' \
          -H 'Connection: keep-alive' \
          -H 'Content-Length: 1034' \
          -H 'Content-Type: multipart/form-data; boundary=--------------------------089094351527063763744770' \
          -H 'Host: localhost:8376' \
          -H 'cache-control: no-cache' \
          -H 'content-type: multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW' \
          -F public_key=@public_key.pem
        ```
    3. Save output above in file. Load encrypted:

       ```
       curl -X POST \
         http://localhost:8376/v1/loadstate \
         -H 'Accept: */*' \
         -H 'Accept-Encoding: gzip, deflate' \
         -H 'Cache-Control: no-cache' \
         -H 'Connection: keep-alive' \
         -H 'Content-Length: 5443' \
         -H 'Content-Type: multipart/form-data; boundary=--------------------------767615378467519380075801' \
         -H 'Host: localhost:8376' \
         -H 'cache-control: no-cache' \
         -H 'content-type: multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW' \
         -F sls_dump=@sls_test_config.json \
         -F private_key=@private_key.pem
       ```
- Liveness/readiness probes.

## [0.6.1]

### Changed

- Changed URLS so they do not begin with /sls/.  The use of /sls insternally was resulting in URLs beginning with /sls/sls/ when transiting the API gateway.

## [0.6.0]

### Added

- Added GET for /hardware (gets list of all hardware components)

## [0.5.0]

### Added

- Added search for hardware and networks.

## [0.4.1]

### Added

- Added /hardware API set

## [0.4.0]

### Added

- Added completed /loadstate and /dumpstate endpoints

## [0.3.0]

### Added

- Added unit testing for database against real database instance.

## [0.2.0]

### Added

- Adds support for all network API operations except for PATCH.

## [0.1.0]

### Added

- This release adds the final bits necessary to support the basic operations of SLS. It also builds and functionally
  runs (though doesn't do anything all that useful yet.)

## [0.0.1]

### Added

- This is the initial release.  It contains no functionality yet.
