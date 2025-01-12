package osutility

import (
	"context"
	"encoding/base64"
	"fmt"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/log/log"
	"os"
	"os/user"
	"strconv"
	"strings"
)

type AuthorizedKeys interface {
	Add(ctx context.Context, ownerName, key string) error
	Delete(ctx context.Context, ownerName string) error
	Update(ctx context.Context, ownerName, key string) error
	Get(ctx context.Context, ownerName string) error
}

func (l LiveAuthorizedKeys) Add(ctx context.Context, ownerName, key string) error {
	logger := log.From(ctx)
	logger.Info("Adding SSH Key", "Owner Name", ownerName)

	// Decode authorized key
	logger.Info("Decoding Key")
	authorisedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		logger.Error(err, "error while decoding key")
		return fmt.Errorf("failed to decode string: %v", err)
	}

	// Lookup administrative user by name
	adminUsername := constants.AdminUserName
	adminUser, err := user.Lookup(adminUsername)
	if err != nil {
		logger.Error(err, "error retrieving information for user %w", adminUsername)
		return fmt.Errorf("error retrieving information for user %v: %v", adminUsername, err)
	}

	// Retrieve UID and GID
	uid := adminUser.Uid
	gid := adminUser.Gid
	UID, err := strconv.ParseInt(uid, constants.Base, constants.BitSize)
	if err != nil {
		logger.Error(err, "error parsing (string)uid to (int)uid")
		return fmt.Errorf("error parsing (string)uid to (int)uid: %v", err)
	}
	GID, err := strconv.ParseInt(gid, constants.Base, constants.BitSize)
	if err != nil {
		logger.Error(err, "error parsing (string)gid to (int)gid")
		return fmt.Errorf("error parsing (string)gid to (int)gid: %v", err)
	}

	// Check if the directory exists, create it if not
	dir := constants.AuthorizedKeyDirPath
	logger.Info("Checking if the directory exists", "DirPath", dir)
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		err := l.fs.MkdirAll(ctx, dir, constants.OwnerReadWriteExecute)
		if err != nil {
			logger.Error(err, "error while creating directory")
			return fmt.Errorf("failed to create directory: %v", err)
		}
		logger.Info("Directory Created", "DirPath", dir)

		// Change ownership of the directory
		err = l.fs.Chown(ctx, dir, int(UID), int(GID))
		if err != nil {
			logger.Error(err, "error while changing directory ownership")
			return fmt.Errorf("error while changing directory ownership: %v", err)
		}
		logger.Info("Ownership changed", "user", adminUsername, "uid", UID, "gid", GID)
	}

	// Check if the file exists, create it if not
	filename := constants.AuthorizedKeyFile
	logger.Info("Checking if file exists", "FilePath", filename)
	var fileExists bool
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logger.Info("File does not exist. Creating!")
		file, err := l.fs.OpenFileWithPermission(ctx, filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, constants.FileReadWriteAccess)
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				logger.Error(err, "error while creating file")
			}
		}(file)
		if err != nil {
			logger.Error(err, "error while creating file")
			return fmt.Errorf("error while creating file: %v", err)
		}
		logger.Info("Created file for adding key", "FilePath", filename)
		fileExists = false

		// Change ownership of the file
		err = l.fs.Chown(ctx, filename, int(UID), int(GID))
		if err != nil {
			logger.Error(err, "error while changing file ownership")
			return fmt.Errorf("error while changing file ownership: %v", err)
		}
		logger.Info("Ownership changed", "user", adminUsername, "uid", UID, "gid", GID)
	} else {
		logger.Info("File already exist.")
		fileExists = true
	}

	// Handle content modification based on whether the file exists
	if fileExists {
		logger.Info("Reading file content")
		content, err := l.fs.ReadFile(ctx, filename)
		if err != nil {
			logger.Error(err, "error while reading file content")
			return fmt.Errorf("failed to read file: %v", err)
		}
		var existingContent = string(content)

		// Check if key is already matched with an owner
		logger.Info("Checking if key is matched with an owner")
		lines := strings.Split(existingContent, "\n")
		keyFound := false

		for _, line := range lines {
			parts := strings.Split(line, " ")
			if len(parts) > 0 && parts[len(parts)-1] == ownerName {
				keyFound = true
				break
			}
		}
		logger.Info("Is key matched with an owner?", "Owner Name", ownerName, "Key Found", keyFound)

		// If key is matched with an owner then delete that entry
		if keyFound {
			logger.Info("Deleting the existing key")
			err = l.fs.DeleteLineFromFileByKey(ctx, filename, ownerName)
			if err != nil {
				logger.Error(err, "error while deleting key")
				return fmt.Errorf("failed to delete key from file: %v", err)
			}
		}
	}

	// Write the new line with key and owner
	logger.Info("Adding new key")
	newContent := fmt.Sprintf("%s %s", string(authorisedKey), ownerName)
	err = l.fs.WriteNewLine(ctx, constants.AuthorizedKeyFile, []byte(newContent))
	if err != nil {
		logger.Error(err, "error while adding new key")
		return fmt.Errorf("failed to write to file: %v", err)
	}
	logger.Info("Key Addition Successful")

	return nil
}

func (l LiveAuthorizedKeys) Delete(ctx context.Context, ownerName string) error {
	// TODO implement me
	panic("implement me")
}

func (l LiveAuthorizedKeys) Update(ctx context.Context, ownerName, key string) error {
	// TODO implement me
	panic("implement me")
}

func (l LiveAuthorizedKeys) Get(ctx context.Context, ownerName string) error {
	// TODO implement me
	panic("implement me")
}

type LiveAuthorizedKeys struct {
	exec Exec
	fs   Filesystem
}

func NewLiveAuthorizedKeys(execUtil Exec, fsUtil Filesystem) *LiveAuthorizedKeys {
	f := &LiveAuthorizedKeys{
		exec: execUtil,
		fs:   fsUtil,
	}

	return f
}
