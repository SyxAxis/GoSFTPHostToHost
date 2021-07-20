# GoSFTPHostToHost

An experiment to test if it's possible to SFTP between two remote hosts, with the local system acting as a "go-between" to handle the data. The reason was to avoid have to run a two-step process, download to local file and then upload. This method allows the local handler to have minimal space as it never has to store anything, it simply needs memory and network bandwidth.
