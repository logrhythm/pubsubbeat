# Building Community Beats

This guide assumes you have a GitHub repo containing a Beat you would like to wrap in a Docker image. If that didn't make sense, or you don't have one, stop now.

## Prereqs
### Repository
This build system works for community Beats that meet the following criteria:
- Includes LICENSE file at root (*not* LICENSE.txt or LICENSE.md)
- Builds with `make`
- Tagged with a sane version number (1.1.0, v2.1.6)

### When Building for Test
See what project you're auth'd to
```bash
gcloud config list
```
When building for test we want to auth to our test project
```bash
gcloud config set project datacollector-215718
```

### When Building for Production
*DO NOT BUILD FOR PRODUCTION LIGHTLY*

See what project you're auth'd to
```bash
gcloud config list
```
When building for prod we want to auth to our prod project
```bash
gcloud config set project lrcollection
```
Finally, revert back to test when finished with the build!
```bash
gcloud config set project datacollector-215718
```

## Determine Substitutions
When running the build we have 2 substitutions:
- Name of the person or organization hosting the repo, in the case of most LogRhythm forked community Beats this name will be `logrhythm`
- Name of the Beat (corresponding to the GitHub repo name, ex: `pubsubbeat`)

**Note:**  
The version is now automatically determined from [latest.yml](https://raw.githubusercontent.com/logrhythm/versions/master/latest.yml) for `pubsubbeat` and prefixed with `dev_`.  
You do **not** need to specify a version substitution.

## Run Build
To run the build from `/pubsubbeat`:
```bash
gcloud builds submit --config ./deploy/beats/cloudbuild.yaml --substitutions=_BEAT_PUBLISHER="publisher",_BEAT_NAME="repo name"
```
*Fill in `publisher` and `repo name` with your values from the previous step!*

GCS Dependencies path -- [logrhythm_datacollector_external_deps](https://console.cloud.google.com/storage/browser/logrhythm_datacollector_external_deps;tab=objects?forceOnBucketsSortingFiltering=false&project=datacollector-215718&prefix=&forceOnObjectsSortingFiltering=false)