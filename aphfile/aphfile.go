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

func FetchRemoteFile(s3Client *minio.Client, bucket, path, name string) (string, error) {
	tmpf, err := ioutil.TempFile("", name)
	if err != nil {
		return "", err
	}
	if err := s3Client.FGetObject(bucket, path, tmpf.Name()); err != nil {
		return "", fmt.Errorf("Unable to retrieve the object %s", err.Error(), 2)
	}
	return tmpf.Name(), nil
}

func Untar(src, target string) error {
	reader, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open file reading %s", err)
	}
	defer reader.Close()
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
		true,
	)
	if err != nil {
		return s3Client, fmt.Errorf("unable create the client %s", err.Error())
	}
	return s3Client, nil
}

func FetchAndDecompress(client *minio.Client, bucket, path, name string) (string, error) {
	filename, err := FetchRemoteFile(client, bucket, path, name)
	if err != nil {
		return "", fmt.Errorf("unable to fetch remote file %s ", err)
	}
	tmpDir, err := ioutil.TempDir(os.TempDir(), name)
	if err != nil {
		return "", fmt.Errorf("unable to create temp directory %s", err)
	}
	err = Untar(filename, tmpDir)
	if err != nil {
		return "", fmt.Errorf("error in untarring file %s", err)
	}
	return tmpDir, nil
}
