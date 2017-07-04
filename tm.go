package mysqlTM

//Package for maintaing a table in mysql. Geared towards account management. Not complete
//but whats here works
import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type TableManager struct {
	//pointer to the mysql database
	db *sql.DB
	//mysql query string
	//"select (verifierField) from (table) where (verifierField) = (verifier)"
	//for verifying things like sign ins, used in tm.Verify()
	verifyQ string
	//mysql query string.
	//"insert into (table)(field, field...)". The "values( ... )" part is added
	//dyanamically right before execution. used in tm.Insert()
	insertQ string
	//holds the amount of fields entered with the constructor (the ...string). Used
	//to verify that enough fields where entered when tm.Insert() is called.
	fieldAmt int
	//name of the table that is being managed. This package could potentially manage
	//many tables, but probably best not to
	tableName string
	//the field name of whatever unique identifier (optional, set with tm.SetupVerify).
	//probably a username or an email.
	unique string
	//the field name of whatever the verifier for the unique name is(also optional ^^).
	//mostlikely a password
	verifier string
	//the fields entered at constructor time
	fields []string
}

//Returns a new TableManager object. Parameters- dbUserName: the username for mysql.
//dbPass: password for that username. host: the host and port you will be connecting to
//(pass empty string "" for the default). db: The name of the database that you're
//connecting to. tableName: the name of the table that will be managed. fields: the
//of that table, as they appear on that table. For now(this sucks), those fields must not
//by any type of number. They must be strings or things that mysql accepts as strings (ie
//datetime, time, varchar, char, etc). Will be changed later to be more dynamic
func NewTM(dbUserName string, dbPass string, host string, db string,
	tableName string, fields ...string) (*TableManager, error) {
	var err error
	tm := new(TableManager)
	sqlopenString := dbUserName + ":" + dbPass + "@" + host + "/" + db
	tm.db, err = sql.Open("mysql", sqlopenString)
	if err != nil {
		return nil, err
	}
	if len(fields) < 1 {
		return nil, errors.New("no tables fields entered")
	}
	tm.fields = fields
	tm.fieldAmt = len(fields)
	tm.tableName = tableName
	tm.insertQ = "insert into " + tableName + "(" + seperateWithCommas(fields, false) + ") "
	tm.verifyQ = "select %v from " + tableName + " where %v = '%v'"
	tm.unique = ""
	tm.verifier = ""
	return tm, nil
}

//Sets up for tm.Verify() to work, otherwise tm.Verify() will return an error. Both
//parameters are the names of the fields as they are in mysql, and as you entered them
//for the constructor (ie: "username" for uniqueFieldName and "password" for
//"verifierFieldName")
func (tm *TableManager) SetupVerify(uniqueFieldName string,
	verifierFieldName string) error {
	//TODO check that they exist in tm.fields
	tm.unique = uniqueFieldName
	tm.verifier = verifierFieldName
	return nil
}

//verifies that the unique field has a matching verifier (username matches password)
//. returns true if it does, false if it doesnt, error if incorrect input was entered
//or something went wrong with the database
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

//Inserts a new entry into the table. fieldsInOrder are the values to be inserted, in
//the order that the fields were entered in the constructor. returns error if something
//went wrong with the database
func (tm *TableManager) Insert(fieldsInOrder ...string) error {

	if len(fieldsInOrder) != tm.fieldAmt {
		return errors.New("Enter a value for all fields")
	}

	//TODO error check the lengths of shit, incase fuckers get clever
	//TODO check if it exists CheckExits() down low
	fullQuery := tm.insertQ + "values(" + seperateWithCommas(fieldsInOrder, true) + ");"
	_, err := tm.db.Exec(fullQuery)
	if err != nil {
		return err
	}
	return nil
}

//idgaf how disgusting this is
//properly seperates shit for the query strings. wrap just wraps shit like this: 'shit'
func seperateWithCommas(strings []string, wrap bool) string {
	var s string
	if wrap {
		s = "'" + strings[0] + "'"
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

/*********************some future functions*************************/
func escapeSequence(str string) string {
	return str
}

//not written
func UpdateEntry(field string, newValue string) error {
	return nil
}

//not written
func (tm *TableManager) CheckExists(uniqueField string, uniqueValue string) (bool, error) {
	return false, nil
}

func encrypt(str string) string {
	return str
}

func decrypt(str string) string {
	return str
}

//not written
func (tm *TableManager) RemoveByUnique(unique string) error {
	return errors.New("this function isnt written")
}

//not written
func (tm *TableManager) Remove(fieldValues ...string) error {
	return errors.New("this function isnt written")
}
