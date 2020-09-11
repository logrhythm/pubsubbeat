# Veracode and GoSec static security scans

## Veracode and GoSec files
* post_vendor.sh     - bash script: post restore 'vendor' directory after veracode and gosec generation
* prep_vendor.sh     - bash script: prepare 'vendor' directory for veracode and gosec generation
* pubsub_go.mod      - automatic generated go.mod file via siem repo's WORKSPACE file's go_repository items
**                   -      veracode_zip.sh uses this file to create veracode zip and gosec report file
* veracode.json      - veracode control json file for pubsubbeat
* veracode_zip.sh    - bash script: generates all component zip files and gosec report files
**                   -      Also supports Prepare and Post Clean

* pubsubbeat repo's veracode branch: pubsub-veraCodeFake
* Changes to these files in pubsubbeat repo, 
** Need to be sync-ed back to siem repo's 'cmd/veraode/pubsubbeat' directory
** siem repo, veracode branch: us7476_veracode

* In pubsubbeat repo's make the folowing directories if they do not exist
** mkdir cmd/veraode
** mkdir cmd/veraode/scripts
** mkdir cmd/veraode/results

* In pubsubbeat repo's make 'cmd/veraode' directory
* Copy these post* prep* pubsub* veracode* files into the pubsubbeat repo's 'cmd/veraode' directory
* Copy sibling 'scripts' directory's gosec*.sh gosec*.py __init__.py files
** to pubsubbeat repo's 'cmd/veracode/scripts' directory