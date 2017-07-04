package mysqlTM
//TODO need to handle sqlinjections and shit
//TODO this package assumes all fields in the mysql table are string related (need ' '
//around the queries)
//TODO need to make/find a package for encrypting shit (maybe crypto)
//custom error type would be awesome
import (
	_"github.com/go-sql-driver/mysql"
	"database/sql"
	"fmt"
	"errors"
)

type TableManager struct {
	db *sql.DB
	verifyQ string
	insertQ string
	fieldAmt int
	tableName string
	unique string
	verifier string
	fields []string
}

func NewTM(dbUserName string, dbPass string, host string, db string, 
						tableName string, fields ...string) (*TableManager, error) {
	var err error
	tm := new(TableManager)
	sqlopenString := dbUserName+":"+dbPass+"@"+host+"/"+db
	tm.db, err = sql.Open("mysql", sqlopenString)
	if err != nil { 
		return nil, err 
	}
	if len(fields) < 1{
		return nil, errors.New("no tables fields entered")	
	}
	tm.fields = fields
	tm.fieldAmt = len(fields)
	tm.tableName = tableName
	tm.insertQ = "insert into "+tableName+"("+seperateWithCommas(fields, false)+") "
	tm.verifyQ = "select %v from "+tableName+" where %v = '%v'"
	tm.unique = ""
	tm.verifier = ""
	return tm, nil

}
func (tm *TableManager) SetupVerify(uniqueFieldName string, 
														verifierFieldName string) error {
	//check that they exist in tm.fields
	tm.unique = uniqueFieldName
	tm.verifier = verifierFieldName
	return nil
}

func (tm *TableManager) Verify(unique string, verifier string) (bool, error) {
				
	if tm.verifier == "" || tm.unique == "" {
		return false, errors.New("Verifier or Unqiue table fields not set. " +
			"Call tm.SetVerifier(verifierFieldName) and tm.SetUnique(uniqueFieldName)")
	}
	var recievedVerifier string
	fullQuery := fmt.Sprintf(tm.verifyQ, tm.verifier, tm.unique, unique)
	err := tm.db.QueryRow(fullQuery).Scan(&recievedVerifier)
	if err != nil {
		return false, err
	}
	return recievedVerifier == verifier, nil

}
//This shit is obviously not very efficient, but how the fuck can you write a completely
//dynamic package that can take in variable amounts of arguments, and still be efficient?
//you gotta do a bunch of work-arounds
//TODO error check the lengths of shit, incase fuckers get clever
func (tm *TableManager) Insert(fieldsInOrder ...string) error{	
	
	if len(fieldsInOrder) != tm.fieldAmt {
		return errors.New("Enter a value for all fields")	
	}
	//check if it exists
	fullQuery := tm.insertQ + "values(" + seperateWithCommas(fieldsInOrder, true) + ");"
	_, err := tm.db.Exec(fullQuery)
	if err != nil{
		return err	
	}
	return nil
}

//idgaf how disgusting this is
func seperateWithCommas(strings []string, wrap bool) string {	
	var s string
	if wrap {
		s = "'"+strings[0]+"'"
	} else {
		s = strings[0]
	}
	for i := 1; i < len(strings); i++ {
		if wrap {
			s += ", '" + strings[i] + "'"
		} else {
			s += ", " + strings[i]	
		}
	}
	return s
}

func Index(strings []string, match string) (int, error) {
	for i, str := range strings {
		if str == match {
			return i, nil
		}	
	}	
	return -1, errors.New("not found")
}
/*********************some future functions*************************/
func escapeSequence (str string) string{
	return str
}

func updateEntry(field string, newValue string) error {
	return nil	
}

func (tm *TableManager) CheckExists(uniqueField string, uniqueValue string) (bool, error) {
	return false, nil			
}

func encrypt(str string) string {
	return str	
}

func decrypt(str string) string {
	return str	
}

func (tm *TableManager) RemoveByUnique(unique string) error {
	return errors.New("this function isnt written")
}

func (tm *TableManager) Remove(fieldValues ...string) error {
	return errors.New("this function isnt written")	
}






