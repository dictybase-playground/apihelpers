package aphfile

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/minio/minio-go"
)

func FetchRemoteFile(s3Client *minio.Client, bucket, path string) (io.Reader, error) {
	return s3Client.GetObject(bucket, path, minio.GetObjectOptions{})
}

func Untar(reader io.Reader, target string) error {
	archive, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("could not read from gzip file %s", err)
	}
	defer archive.Close()
	tarReader := tar.NewReader(archive)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(file, tarReader)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unable to figure out the file type of tar archive %s %s", header.Typeflag, err)
		}
	}
	return nil
}

func GetS3Client(server, access, secret string) (*minio.Client, error) {
	s3Client, err := minio.New(
		server,
		access,
		secret,
		false,
	)
	if err != nil {
		return s3Client, fmt.Errorf("unable create the client %s", err.Error())
	}
	return s3Client, nil
}

func FetchAndDecompress(client *minio.Client, bucket, path, name string) (string, error) {
	reader, err := FetchRemoteFile(client, bucket, path)
	if err != nil {
		return "", fmt.Errorf("unable to fetch remote file %s ", err)
	}
	tmpDir, err := ioutil.TempDir(os.TempDir(), name)
	if err != nil {
		return "", fmt.Errorf("unable to create temp directory %s", err)
	}
	err = Untar(reader, tmpDir)
	if err != nil {
		return "", fmt.Errorf("error in untarring file %s", err)
	}
	return tmpDir, nil
}
