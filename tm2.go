//TODO to fucking do, all these string concatenations im doing might be unecessary, db.Exec
//and db.Query all take parameters as options. Maybe not entirely unecessary, but slightly
//select * from table where something = ? or someshit
//TODO in fucking fact, the ?ing handles the '' shit for you i think, so thatll make the
//multiple typing a breeze!
package mysqlTM
import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
)

type field stuct {
	field string
	ttype string
	null string
	key string
	ddefault interface{}
	extra string
}

type TableManager struct {
	db *sql.DB
	verifyQ string
	insertQ string
	existsQ string
	fieldAmt int
	tableName string
	unique string
	verifier string
	fields []field
	verifySet bool
}

//NUG BUG
func NewTM(dbUserName string, dbPass string, host string, db string, 
												tableName string) (*TableManager, error) {
	var err error
	tm := new(TableManager)
	sqlopenString := dbUserName + ":" + dbPass + "@" + host + "/" + db
	
	tm.db, err = sql.Open("mysql", sqlopenString)
	if err != nil {
		return nil, err
	}
	tm.verifySet = false
	tm.tableName = tableName
	tm.verifyQ = "select %v from " + tableName + " where %v = '%v'"
	tm.unique = ""
	tm.verifier = ""
	err = tm.getFields()
	if err != nil {
		return nil, err	
	}
	tm.fieldAmt = len(tm.fields)
	//BUG BUG figure this out, this is not longer just a slice of string fucker BUG BUG
	tm.insertQ = "insert into " + tableName + "(" + seperateWithCommas(fields, false)+") "
	return tm, nil
}

func (tm *TableManager) SetupVerify(uniqueFieldName string,
									verifierFieldName string) error {
	tm.unique = uniqueFieldName
	tm.verifier = verifierFieldName
	tm.existsQ = "select * from "+tm.tableName+" where "+uniqueFieldName+" = "
	tm.verifySet = true
	return nil
}

func (tm *TableManager) Verify(unique string, verifier string) (bool, error) {

	if !tm.verifySet{
		return false, errors.New("tm.SetVerify not called")	
	}
	var recievedVerifier string
	fullQuery := fmt.Sprintf(tm.verifyQ, tm.verifier, tm.unique, unique)
	err := tm.db.QueryRow(fullQuery).Scan(&recievedVerifier)
	return recievedVerifier == verifier, err

}

//NUG BUG probably janky
func (tm *TableManager) InsertR(r *http.Request) error {
	values, err := getValuesFromForm(r)
	if err != nil {
		return err	
	}
	return tm.Insert(values...)
}

func (tm *TableManager) Insert(fieldsValuesInOrder ...string) error {

	if len(fieldValuesInOrder) != tm.fieldAmt {
		return errors.New("Enter a value for all fields")
	}

	fullQuery := tm.insertQ + "values("+seperateWithCommas(fieldValuesInOrder, true)+");"
	_, err := tm.db.Exec(fullQuery)
	return err
}

//NUG BUG
//TODO all of these will be ...interface{}
//Testing this shit af
func (tm *TableManager) InsertPartial(fieldsAndValues ...string) {
	//This block of code is getting the proper query string... TODO i need to find a more
	//efficient way of doing this... this, along with GetAll, is just fucked
	l := len(fieldsAndValues)
	if l % 2 != 0 || l == 0{
		return nil,errors.New("should be GetAll(field1, value1, field2, value2). read doc")
	}
	q := "insert into "+tm.tableName
	fieldsPartialQ := "(" + fieldsAndValues[i]
	valuesPartialQ := "values(?"
	values := make([]interface{}, l/2)
	values[0] = fieldsAndValues[1]
	for i := 2; i < l; i += 2 { 
		fieldsPartialQ += "," +fieldsAndValues[i]
		valuesPartialQ += ", ?"
		values[i/2] = fieldsAndValues[i]
	}
	q += fieldsPartialQ + ") "+ valuesPartialQ +"?;"
	
	
	_, err := tm.db.Exec(q, values)
	return err
}

//NUG BUG look over
func (tm *TableManager) GetByUnique(value string)([]string, error) {
	if !tm.verifySet {
		return nil, errors.New("Need to call SetupVerify for unique to be set")	
	}
	row, err := tm.db.Query("select * from "+tm.tableName+" where "+tm.unique+" = "+value
	if err != nil {
		return nil, err	
	}
	entry := make([]string, tm.fieldAmt)
	if row.Next(){
		err = row.Scan(entry...)
		return entry, err 
	}					
	return nil, erros.New("No entry with value "+value+" for field "+tm.unique)
}

//this is fucked TODO
//also BUG
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

func (tm *TableManager) CheckExists(uniqueValue string) (bool, error) {
	if tm.existsQ == "" {
		return true, errors.New("tm.SetupVerify not called")	
	}

	fullQ := tm.existsQ + "'" + uniqueValue + "'"
	fmt.Println(fullQ) 
	rows, err := tm.db.Query(fullQ)
	if err != nil {
		return true, err	
	}
	if rows.Next(){
		return true, nil	
	}
	return false, nil
}

func escapeSequence(str string) string {
	return str
}

func encrypt(str string) string {
	return str
}

func decrypt(str string) string {
	return str
}

//TODO itd be nice to not be allocating memory everywhere for these. Maybe see if arrays
//work as well?
func (tm *TableManager) DeleteR(r *http.Request) error {
	values, err := getValuesFromForm(r)
	if err != nil {
		return err	
	}
	return tm.Delete(values...)
}

//BUG
func (tm *TableManager) Delete(fieldValuesInOrder ...string) error {
	if len(fieldValuesInOrder != tm.fieldAmt){
		return errors.New("Incorrect field amounts entered")	
	}
	fullQ := //someshit prolls tm.deleteQ + someshit
	_, err := tm.db.Exec(fullQ)
	return err
}

//NUG BUG
func (tm *TableManager) DeleteByUnique(value string) error {
	if !tm.verifySet {
		return nil, errors.New("Need to call SetupVerify for unique to be set")	
	}
	fullQ := fmt.Sprintf(tm.deleteUniqueQ, value)
	_, err := tm.db.Exec(fullQ)
	return err	
}

//NUG BUG probably janky
func (tm *TableManager) getFields() error {
	rows, err := tm.db.Query("describe " + tm.tableName)	
	if err != nil {
		return erros.New("Error when checking table :" + err.Error())	
	}
	fields := make([]field, 0)
	var r field
	for rows.Next() {
		err := rows.Scan(&r.Field, &r.Type, &r.Null, &r.Key, &r.Default, &r.Extra)
		if err != nil{
			return errors.New("Error during scanning :", err.Error())	
		}
		fields = append(fields, r)
	}
	tm.fields = fields
	return nil
}

//NUG BUG
func (tm *TableManager) getValuesFromForm(r *http.Request) ([]string, error) {
	r.ParseForm()
	values := make([]string, tm.fieldAmt)	
	for i, field := tm.fields {
		val, ok := r.Form(field.field)
		if !ok {
			return nil, errors.New("Form from *http.Request doesnt have "+field.field)	
		}
		values[i] = val
	}
	return values, nil
}

//NUG BUG
func (tm *TableManager)UpdateByUnique(uniqueVal string, field string, newVal string) error{
	if !tm.verifySet {
		return nil, errors.New("Need to call SetupVerify for unique to be set")	
	}
	fullQ := fmt.Sprintf(tm.updateQ, uniqueVal, field, newVal)
	_, err := tm.db.Exec(fullQ)
	return err
}

//NUG BUG idk lol this is really macheteed. 
//TODO once all table types are allowed, this will return [][]interface{}
//TODO do GetAll(max int, fieldsAndValues ...string) that will get a maximum number
func (tm *TableManager) GetAll(fieldsAndValues ...string)([][]string, error) {
	//TODO this get query section is pretty bad, but it seems to work. I knows its a
	//fucking mess, but all it does is get the query to be ran in the form of
	//"select * from table where field = value and field2 = value2 and field3 = ..."
	//dynamically, based on how many field and value pairs were entered
	l := len(fieldsAndValues)
	if l % 2 != 0 {
		return nil,errors.New("should be GetAll(field1, value1, field2, value2). read doc")
	}
	q := "select * from "+tm.tableName
	//this allows for function to be passed nothing, which means get everything from table
	if l != 0 {
		q += " where "+fieldsAndValues[0]+" = '" + fieldsAndValues[1]+"' "
	}
	for i := 2; i < l; i += 2 { //TODO that '%v' is hardcoded for strings
		q += "and " + fmt.Sprintf("%v = '%v'", fieldsAndValues[i], fieldsAndValues[i+1])
	}
	
	rows, err := tm.db.Query(q)
	if err != nil {
		return nil, err	
	}
	//unfortunately, no way to do this without append, since theres no way to tell how
	//many rows are returned. TODO sucks that i have to allocate so much fucking memory.
	//see what brad fitzpatricks comment in .Next() or .Scan() was about non-allocated, 
	//read only memory or someshit
	//see if theres a way to know ahead of time how many entries in rows
	entries := make([][]string, 0) // will eventually be [][]interface{}
	for rows.Next() {
		entry := make([]string, tm.fieldAmt)
		err = rows.Scan(entry...)
		if err != nil {
			return nil, err	
		}
		entries := append(entries, entry)
	}
	if err = rows.Err(); err != nil {
		return nil, err	
	}
	return entries, nil
}

//These are just "not enough behavior in this package" functions
func (tm *TableManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tm.db.Query(query, args...)	
}

func (tm *TableManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tm.db.Exec(query, args...)	
}







