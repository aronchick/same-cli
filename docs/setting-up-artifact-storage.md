After each build, we upload a SAME binary to an Azure Blob so that people can download it.

We're following these instructions: https://docs.microsoft.com/en-us/azure/storage/blobs/storage-blobs-static-site-github-actions

To do so:
- We copy the appropriate files to a temporary directory
- Generate a signature for the binary file
- Then upload the binary and the signature to the Azure blob
- And upload a sym link from latest to point at the new upload
  
There is already an Azure function which verifies the signature, downloads the file, and then reverifies the download before moving it to /usr/local/bin