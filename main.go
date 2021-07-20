package main

/*

	Auth: G. Johnson
	Date: 20-JUL-2021
	Desc: Tramsfer files direct between 2 remote SFTP hosts, acting as "go-between"

	SFTP lib uses SSH keys, use Puttygen to generate the pub/priv key pair files and  make sure to save priv key as "OpenSSH" format ( NOT new OpenSSh )

*/

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type ConnectionConfig struct {
	SFTPhost           string
	SFTPport           string
	SFTPuser           string
	SFTPprivatekeyfile string
}

func checkErr(err error, msg string) {

	if err != nil {
		log.Fatal("FAILED: ", msg)
		log.Fatal("ERROR : ", err)
	}
}

func GetSFTPConnection(connConfigSettings *ConnectionConfig) *sftp.Client {

	// pull in the private key file data
	pemBytes, err := ioutil.ReadFile(connConfigSettings.SFTPprivatekeyfile)
	checkErr(err, "Unable to find private key file "+connConfigSettings.SFTPprivatekeyfile)

	// parse the private key to make sure it's sound
	signer, err := ssh.ParsePrivateKey(pemBytes)
	checkErr(err, "Unable to parse the private keyfile.")

	// attach using a private key and a valid hostkey string
	// Note: INSECURE connection ignoring the remote host's hostkey
	//       will give your sec admin a fit!!
	config := &ssh.ClientConfig{
		User: connConfigSettings.SFTPuser,
		// alterntaive is to use a plain text password
		// Auth: []ssh.AuthMethod{ ssh.Password(*flgPassword),
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		// ignore the hostkey and just accept it, NOT a good idea in prod/live envs
		// especially on the nasty internet!
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// establish the SSH connection. SFTP is really just FTP over an SSH tunnel
	// open the SSH tunnel
	fmt.Printf("Conn [%v]\n", connConfigSettings.SFTPhost)
	conn, err := ssh.Dial("tcp", connConfigSettings.SFTPhost+":"+connConfigSettings.SFTPport, config)
	checkErr(err, "Unable to open the remote connection.")

	// once the ssh hooked up, attach the SFTP connection down the SSH tunnel
	client, err := sftp.NewClient(conn)
	checkErr(err, "Unable to secure a remote SFTP client connection.")

	// get the connection back
	return client

}

func TransferFile(sftpSvr01, sftpSvr02 *sftp.Client, srcFile, tgtFile string) error {

	// open the source file from the first server
	rmtFile01, err := sftpSvr01.Open(srcFile)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer rmtFile01.Close()

	// open a write stream to the target host/file
	rmtFile02, err := sftpSvr02.Create(tgtFile)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer rmtFile02.Close()

	// start shifting the bytes
	// love go, "pull the data into the target"!
	bytes, err := io.Copy(rmtFile02, rmtFile01)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Printf("\n[%d] bytes copied\n", bytes)

	return nil

}

func main() {

	// spec up to the two paths/files
	srcFile := "/home/gxj/file_src.txt"
	tgtFile := "/home/gxj/file_tgt.txt"

	// define the first connection
	clientConfig01 := &ConnectionConfig{
		SFTPhost:           "192.168.56.123",
		SFTPport:           "22",
		SFTPuser:           "user",
		SFTPprivatekeyfile: "C::\\Dev Work\\golang\\sftp_ssh1.ppk.txt",
	}

	// define the second connection
	// yes this can be the same host as the first host if you like!
	clientConfig02 := &ConnectionConfig{
		SFTPhost:           "192.168.56.123",
		SFTPport:           "22",
		SFTPuser:           "user",
		SFTPprivatekeyfile: "C::\\Dev Work\\golang\\sftp_ssh1.ppk.txt",
	}

	// get the connections open
	fmt.Println("Conn start...")
	sftpSvr01 := GetSFTPConnection(clientConfig01)
	defer sftpSvr01.Close()

	sftpSvr02 := GetSFTPConnection(clientConfig02)
	defer sftpSvr02.Close()

	fmt.Printf("Copying [%v]:[%v] to [%v]:[%v]....", clientConfig01.SFTPhost, srcFile, clientConfig02.SFTPhost, tgtFile)

	// get the transfer going
	err := TransferFile(sftpSvr01, sftpSvr02, srcFile, tgtFile)
	checkErr(err, "Transfer has failed!")

}
