package pwsafe

import "os"

//OpenPWSafeFile Opens a password safe v3 file and decrypts with the supplied password.
//Records is keyed by Title (see V3.KeyByUUID); use OpenPWSafeFileKeyedByUUID to key by
//UUID instead, which retains every record in databases with duplicate titles.
func OpenPWSafeFile(dbPath string, passwd string) (*V3, error) {
	return openPWSafeFile(dbPath, passwd, false)
}

//OpenPWSafeFileKeyedByUUID Opens a password safe v3 file and decrypts with the supplied
//password, keying the resulting V3.Records by UUID instead of Title. Password Safe v3
//only guarantees UUID to be unique, not Title, so prefer this over OpenPWSafeFile when
//the database may contain multiple records sharing a title (e.g. several "Google"
//entries) to avoid having all but one of them silently dropped.
func OpenPWSafeFileKeyedByUUID(dbPath string, passwd string) (*V3, error) {
	return openPWSafeFile(dbPath, passwd, true)
}

func openPWSafeFile(dbPath string, passwd string, keyByUUID bool) (*V3, error) {
	var db V3
	db.KeyByUUID = keyByUUID

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
