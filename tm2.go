package main

import (
	"database/sql"
	"errors"
	"fmt"
	//"net/http"
	//crypto
	_ "github.com/go-sql-driver/mysql"
)

//TODO custom error types, so you can do outside package something like
//if err == mysqlTM.VerifyNotSet
//for if they dont do verify or someshit

type field struct {
	field string
	ttype string
	null string
	key string
	//in wherever this is used, deafult is tried to be assigned the value <nil>, but
	//strings cant be nil. macheteed
	ddefault interface{}
	extra string
}

type TableManager struct {
	db *sql.DB
	verifyQ string
	insertQ string
	deleteQ string
	deleteUniqueQ string 
	uniqueExistsQ string
	updateByUniqueQ string
	getByUniqueQ string
	fieldAmt int
	table string
	unique string
	verifier string
	fields []field
	uniqueSet bool
}

func NewTM(dbUserName string, dbPass string, host string, db string, 
												table string) (*TableManager, error) {
	var err error
	tm := new(TableManager)
	sqlopenString := dbUserName + ":" + dbPass + "@" + host + "/" + db
	
	tm.db, err = sql.Open("mysql", sqlopenString)
	if err != nil {
		return nil, err
	}

	tm.uniqueSet = false
	tm.table = table
	
	if tm.getFields() != nil  {
		return nil, err	
	}

	tm.fieldAmt = len(tm.fields)
	tm.setInsertQ()
	tm.setDeleteQ()
	return tm, nil
}

//sets the unique field and the verifier field(ie: username and password, respectively). 
//also runs some set up for future use
//sets for the duration of executaion
func (tm *TableManager) SetUnique(uniqueField string, verifierField string) error {
	tm.unique = uniqueField
	tm.verifier = verifierField
	
	tm.setUniqueExistsQ()
	tm.setDeleteUniqueQ()
	tm.setVerifyQ()
	tm.setUpdateByUniqueQ()
	tm.setGetByUniqueQ()

	tm.uniqueSet = true
	return nil
}

//Verifies that the given unique value has the given verifier (ie: testing someones log in)
//returns true if they match. Returns an error if something went wrong in the db
func (tm *TableManager) Verify(unique string, verifier string) (bool, error) {

	if !tm.uniqueSet{
		return false, errors.New("tm.SetUniwue not called")	
	}
	var recievedVerifier string
	err := tm.db.QueryRow(tm.verifyQ, unique).Scan(&recievedVerifier)
	return recievedVerifier == verifier, err
}
/*
func (tm *TableManager) InsertR(r *http.Request) error {
	values, err := getValuesFromForm(r)
	if err != nil {
		return err	
	}
	return tm.Insert(values...)
}
*/

//Takes the field values in the order that they appear in the mysql table. Does not check
//if the unique value already exists. tm.AlreadyExists should be called before this 
//returns an err if something went wrong with database driver
func (tm *TableManager) Insert(fieldValuesInOrder ...interface{}) error {

	if len(fieldValuesInOrder) != tm.fieldAmt {
		return errors.New("Enter a value for all fields")
	}
	_, err := tm.db.Exec(tm.insertQ, fieldValuesInOrder...)
	return err
}

/*
//Testing this shit af
func (tm *TableManager) InsertPartial(fieldsAndValues ...string) {
	//This block of code is getting the proper query string... TODO i need to find a more
	//efficient way of doing this... this, along with GetAll, is just fucked
	l := len(fieldsAndValues)
	if l % 2 != 0 || l == 0{
		return nil,errors.New("should be GetAll(field1, value1, field2, value2). read doc")
	}
	q := "insert into "+tm.table
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
*/

//Gets all fields of an entry with that unique value. Returns them as a slice of
//interface{}, in order as they are in mysql table, otherwise an error if something went
//wrong. Need to call SetUnique sometime before this
func (tm *TableManager) GetByUnique(unique interface{})([]interface{}, error) {
	if !tm.uniqueSet {
		return nil, errors.New("Need to call SetupVerify for unique to be set")	
	}

	//https://stackoverflow.com/questions/17845619/how-to-call-the-scan-variadic-function-in-golang-using-reflection 
	//hacked af. Scan requires addres of variables, so thats what going on with these
	//first to makes and the forloop. Also, Scan is returning a slice of bytes for each val
	//. since that isnt useful, and you cant just type cast something of type interface{},
	//we loop over the fields and do a switch of the field.ttype, and depending on what 
	//that is, do an appropriate type assertion (which i think (i could be wrong) is what 
	//that .([]byte)) is. link should be a little bit of guidance. Not much tho. Also think
	//of a better way to do this. but for now, TODO add in the other types to the swith
	//that sql has so this returns the appropriate things

	values := make([]interface{}, tm.fieldAmt)
	entry := make([]interface{}, tm.fieldAmt)
	for i, _ := range entry {
		entry[i] = &values[i]	
	}

	err := tm.db.QueryRow(tm.getByUniqueQ, unique).Scan(entry...)
	
	for i, field := range tm.fields {
		switch field.ttype {
			case "int":
			case "whatever":
			default:
				b, ok := values[i].([]byte)
				if ok {
					values[i] = string(b)
				}	
		}	
	}
	return values, err
}

//checks if the unique value set by SetUnique exists. Returns true if it does, false other
//wise. returns err if if something went wrong with the database
func (tm *TableManager) CheckUniqueExists(value interface{}) (bool, error) {
	if !tm.uniqueSet {
		return true, errors.New("tm.SetupVerify not called")	
	}
	rows, err := tm.db.Query(tm.uniqueExistsQ, value)
	defer rows.Close()
	if err != nil {
		return false, err	
	}
	return rows.Next(), rows.Err() 
}

/*

func (tm *TableManager) DeleteR(r *http.Request) error {
	values, err := getValuesFromForm(r)
	if err != nil {
		return err	
	}
	return tm.Delete(values...)
}
*/

//Takes the values of the fields, in order as they appear in mysql table. All fields must
//match for it to be deleted. Returns an err if something went wrong
func (tm *TableManager) Delete(fieldValuesInOrder ...interface{}) error {
	if len(fieldValuesInOrder) != tm.fieldAmt {
		return errors.New("Incorrect field amounts entered")	
	}
	_, err := tm.db.Exec(tm.deleteQ, fieldValuesInOrder...)
	return err
}

//deletes by the unique field. Have to call SetUnique before. Returns error if something
//went wrong with the database
func (tm *TableManager) DeleteByUnique(value interface{}) error {
	if !tm.uniqueSet {
		return errors.New("Need to call SetupVerify for unique to be set")	
	}
	_, err := tm.db.Exec(tm.deleteUniqueQ, value)
	return err	
}

/*
//this will set the values to whatever is in the form from the request. They name of 
//the <input> in html MUST be the same as the name of the field in mysql. If not, this
//wont retrieve that from the form, and will set the value in the mysql field to an empty 
//string. 
func (tm *TableManager) getValuesFromForm(r *http.Request) ([]interface{}, error) {
	r.ParseForm()
	values := make([]interface{}, tm.fieldAmt)	
	//r.Form.Get only returns strings. Empty string if the field.field is not in the form.
	//this is obviously a problem, since even if the field is, lets say, a number, then
	//it will return a "6" instead of 6. fix later if necessary (which it might be) TODO
	//use field.type to know when to convert to number. This will hinder the ability for
	//things to not have to be all string types
	for i, field := tm.fields {
		val := r.Form.Get(field.field)
		values[i] = val
	}
	return values, nil
}
*/

//update the field of the unique value passed to the newVal. Need to call uniqueSet at 
//some point before
func (tm *TableManager)UpdateByUnique(unique interface{}, field string, 
															newVal interface{}) error{
	if !tm.uniqueSet {
		errors.New("Need to call SetupVerify for unique to be set")	
	}
	q := fmt.Sprintf(tm.updateByUniqueQ, field)
	_, err := tm.db.Exec(q, newVal, unique)
	return err
}

/*
//this got fucked by some copy and paste
//TODO do GetAll(max int, fieldsAndValues ...string) that will get a maximum number
func (tm *TableManager) GetAll(fieldsAndValues ...string)([][]string, error) {
	// this get query section is pretty bad, but it seems to work. I knows its a
	//fucking mess, but all it does is get the query to be ran in the form of
	//"select * from table where field = value and field2 = value2 and field3 = ..."
	//dynamically, based on how many field and value pairs were entered
	l := len(fieldsAndValues)
	if l % 2 != 0 {
		return nil,errors.New("should be GetAll(field1, value1, field2, value2). read doc")
	}
	q := "select * from "+tm.table
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
	entries := make([][]string, 0) 
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

*/

//These are just "not enough behavior in this package" functions
func (tm *TableManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tm.db.Query(query, args...)	
}

//for building on top 
func (tm *TableManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tm.db.Exec(query, args...)	
}


/**********************************unexported functions***********************************/


//calls "describe table" to mysql, which returns row of each field in the table
//store them in field struct, then append them to the slice of fields in tm.fields
//no real way to tell how many results are in the sql.Rows object, so have to start by
//making the fields slice at len 0 and continually appending
func (tm *TableManager) getFields() error {
	rows, err := tm.db.Query("describe " + tm.table)	
	if err != nil {
		return errors.New("Error when checking table :" + err.Error())	
	}
	fields := make([]field, 0)
	var f field
	for rows.Next() {
		err := rows.Scan(&f.field, &f.ttype, &f.null, &f.key, &f.ddefault, &f.extra)
		if err != nil{
			return errors.New("Error during scanning :" + err.Error())	
		}
		//fmt.Println("getFields():", f.ttype)
		fields = append(fields, f)
	}
	tm.fields = fields
	return nil
}

//sets the query string for CheckUniqueExists. called from SetUnique
func (tm *TableManager) setUniqueExistsQ() {
	tm.uniqueExistsQ = "select * from "+tm.table+" where "+tm.unique+" = ?"	
}

//sets the queur for the Insert function, since it will always be the same. form of
//"insert into table(field, field) values(?, ?)". question marks because sql.Exec will 
//handle that if you pass it the parameters as well. convenient
//doesnt work out if the table has 0 rows. called after getFields()
func (tm *TableManager) setInsertQ() {
	q := "insert into "+tm.table+"("+tm.fields[0].field
	qMarks := "values(?"
	for i := 1; i < tm.fieldAmt; i++ {
		q += ", " + tm.fields[i].field
		qMarks += ", ?"
	}
	tm.insertQ = q + ") " + qMarks + ");"
}

//sets the query string for delete as "Delete from table where field = ? and field = ?"
//dynamic, called after getFields(). ? again becasue sql.Exec takes care of concatenating
//if you pass parameters
func (tm *TableManager) setDeleteQ() {
	q := "delete from "+tm.table+" where "+tm.fields[0].field+" = ?"
	for i := 1; i < tm.fieldAmt; i++ {
		q += " and "+tm.fields[i].field +" = ?"	
	}	
	tm.deleteQ = q
}

//sets deleteUnique query for DeleteByUnique
//could probably have merged this and setDelete but nah this is easier
func (tm *TableManager) setDeleteUniqueQ() {
	tm.deleteUniqueQ = "delete from "+tm.table+" where "+tm.unique+" = ?"
}

//sets the query for Verify. called by SetVerify
func (tm *TableManager) setVerifyQ() {
	tm.verifyQ = "select "+tm.verifier+" from "+tm.table+" where "+tm.unique+" = ?"
}

//sets the updateByUniqueQ
func (tm *TableManager) setUpdateByUniqueQ() {
	tm.updateByUniqueQ = "update "+tm.table+" set %v = ? where "+tm.unique+" = ?"
}

//setts the get by unique query
func (tm *TableManager) setGetByUniqueQ() {
	tm.getByUniqueQ = "select * from "+tm.table+" where "+tm.unique+" = ?"	
}
/*
func escapeSequence(str string) string {
	return str
}

func encrypt(str string) string {
	return str
}
*/

