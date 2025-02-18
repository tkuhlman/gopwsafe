package pwsafe

import "os"

//OpenPWSafeFile Opens a password safe v3 file and decrypts with the supplied password
func OpenPWSafeFile(dbPath string, passwd string) (*V3, error) {
	var db V3

	// Open the file
	f, err := os.Open(dbPath)
	if err != nil {
		return &db, err
	}
	defer f.Close()

	_, err = db.Decrypt(f, passwd)

	db.LastSavePath = dbPath

	return &db, err
}

//WritePWSafeFile Writes a pwsafe.DB to disk, using either the specified path or the LastSavedPath
func WritePWSafeFile(v3db *V3, path string) error {
	var savePath string
	if path == "" {
		savePath = v3db.LastSavePath
	} else {
		savePath = path
		v3db.LastSavePath = path
	}
	// Open the file
	f, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return v3db.Encrypt(f)
}
