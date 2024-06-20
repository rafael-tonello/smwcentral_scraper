package misc

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
)

type PersistentVarStorage struct {
	location string
}

func NewPersistentVarStorage(location string) PersistentVarStorage {
	// Create a new persistent variable with a filename

	//deteminte the real path of the location (use pwd command)
	currentlocation, _ := os.Getwd()
	location = path.Join(currentlocation, location)

	os.MkdirAll(location, os.ModePerm)
	ret := PersistentVarStorage{location: location}
	return ret
}

func (pv *PersistentVarStorage) Set(key string, value DynamicVar) error {
	// Set the value of a persistent variable
	filename := createFileName(pv.location, key)

	// Write the value to the file
	file, err := os.Create(filename)
	if err != nil {
		// handle error
		return DerivateError(err, "Error creating file")
	}
	defer file.Close()

	_, err = file.WriteString(value.GetString())
	if err != nil {
		// handle error
		return DerivateError(err, "Error writing to file")
	}

	return nil

}

func (pv *PersistentVarStorage) Get(key string) (DynamicVar, error) {
	// Get the value of a persiste	n	// Create a filename for a persistent variable
	filename := createFileName(pv.location, key)

	// Read the value from the file
	file, err := os.Open(filename)
	if err != nil {
		// handle error
		return NewEmptyDynamicVar(), DerivateError(err, "Error opening file")
	}
	defer file.Close()

	// Read the value from the file
	value, err := ioutil.ReadAll(file)
	if err != nil {
		// handle error
		return NewEmptyDynamicVar(), DerivateError(err, "Error reading from file")
	}

	return NewDynamicVar(WithString(string(value))), nil
}

func (pv *PersistentVarStorage) GetOrDefault(key string, defaultValue DynamicVar) (DynamicVar, error) {
	// Get the value of a persistent variable or return a default value
	value, err := pv.Get(key)
	if err != nil {
		return defaultValue, err
	}

	return value, nil
}

func (pv *PersistentVarStorage) IncOrDecInt(key string, amount int64) (DynamicVar, error) {
	// Increment an integer persistent variable
	value, err := pv.GetOrDefault(key, NewDynamicVar(WithInt(0)))

	intValue, _ := value.GetInt64e()

	intValue += amount

	value.SetInt64(intValue)

	err = pv.Set(key, value)

	return value, err
}

func (pv *PersistentVarStorage) SearchVars(searchPattern string) []string {
	// Search for variables in the persistent variable storage
	files, err := os.ReadDir(pv.location)
	if err != nil {
		// handle error
		return []string{}
	}

	ret := []string{}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		//check using regex
		fName := f.Name()
		matched, err := regexp.MatchString(searchPattern, fName)
		if err == nil {
			if searchPattern == "" || matched {
				ret = append(ret, fName)
			}
		}
	}

	return ret
}

func createFileName(location string, key string) string {
	return path.Join(location, key)
}
