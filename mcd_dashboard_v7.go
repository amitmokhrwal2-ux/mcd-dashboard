// mcd_dashboard_v11.1.go
// Live Mode ON. Fixes from user feedback:
// - Robust EmployeeID normalization across Basic/Services (digits-only, strip .0, trim)
// - Better SchoolID extraction from "School Name & ID"
// - Category map recognizes UR/GENERAL/GEN as same
// - DOB/DOJ parsing tolerant to month case + more formats; computes Age reliably
// - Proper counters for Schools/Zones/Designations
// - Marital statuses and categories populated into EMP
// - Added debug logs for joins, DOB/DOJ parsed
// Build & run:
//
//	go run mcd_dashboard_v11.1.go \
//	  -basic ./Basic.csv \
//	  -services ./Services.csv \
//	  -dbt ./Dashboard_Summary_202509.csv \
//	  -template ./mcd_dashboard_template_v11.1.html \
//	  -out ./out/mcd_dashboard_202509_v11.1.html \
//	  -cache ./out/doj_cache.json \
//	  -rebuild
//
// Then open http://localhost:8080
package main

import (
	"bufio"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	basicCSV     = flag.String("basic", "Basic.csv", "Basic CSV path")
	servCSV      = flag.String("services", "Services.csv", "Services CSV path")
	dbtCSV       = flag.String("dbt", "Dashboard_Summary_202509.csv", "DBT CSV path")
	tplPath      = flag.String("template", "mcd_dashboard_template_v11.1.html", "HTML template path")
	outHTML      = flag.String("out", "./out/mcd_dashboard_202509_v11.1.html", "Output HTML path")
	title        = flag.String("title", "MCD Dashboard", "Document title")
	listen       = flag.String("listen", ":8080", "HTTP listen address")
	cachePath    = flag.String("cache", "./out/doj_cache.json", "DOJ cache path (optional)")
	forceRebuild = flag.Bool("rebuild", false, "Force rebuild DOJ cache")
	EmailFrom    = "you@example.com"
	EmailPass    = "xxxxxxxxxxxxxxxx"
	EmailName    = "HQ Team IT Education"
	LiveMode     = true // per user request
)

var (
	bRows, sRows, dRows [][]string
	bm, sm, dm          map[string]int
)

var (
	emailByEmpID   = map[string]string{}
	nameByEmpID    = map[string]string{}
	genderByEmpID  = map[string]string{}
	catByEmpID     = map[string]string{} // from Services.csv
	maritalByEmpID = map[string]string{} // from Basic.csv preferred, else Services.csv (if present)
)

type Emp struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Designation       string `json:"designation"`
	DOB               string `json:"dob"`
	Gender            string `json:"gender"`
	Zone              string `json:"zone"`
	SchoolID          string `json:"school_id"`
	SchoolName        string `json:"school_name"`
	Status            string `json:"status"`
	SelectionCategory string `json:"selection_category"`
	MaritalStatus     string `json:"marital_status"`
	Age               int    `json:"age"`

	Mobile string `json:"mobile"`
	Email  string `json:"email"`
	DOJ    string `json:"doj"`
}
type SchoolDemo struct {
	Zone  string `json:"zone"`
	Count int    `json:"count"`
}

type School struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Zone       string `json:"zone"`
	SIName     string `json:"si_name"`
	MaxTotal   int    `json:"max_total_students"`
	MaxPresent int    `json:"max_present_students"`
	WithAcc    int    `json:"with_account"`
	WithoutAcc int    `json:"without_account"`
	WithAad    int    `json:"with_aadhaar"`
	WithoutAad int    `json:"without_aadhaar"`
	AadLinkAC  int    `json:"aadhaar_linked_account"`
	NewMon     int    `json:"new_admission_this_month"`
	NewSess    int    `json:"new_admission_this_session"`
	DBTStud    int    `json:"dbt_received_student"`
	DBTPar     int    `json:"dbt_received_parent"`
	RecSP      int    `json:"received_by_student_parent"`
}

type SchoolStaff struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Zone           string  `json:"zone"`
	NeededTeachers int     `json:"needed_teachers"`
	ActualTeachers int     `json:"actual_teachers"`
	SurplusVacancy int     `json:"surplus_vacancy"`
	HasPrincipal   bool    `json:"has_principal"`
	HasSpecialEdu  bool    `json:"has_special_educator"`
	TotalStaff     int     `json:"total_staff"`
	Ratio          float64 `json:"ratio"`
}

type DemoStats struct {
	GenMale     int `json:"gen_male"`
	GenFemale   int `json:"gen_female"`
	SCMale      int `json:"sc_male"`
	SCFemale    int `json:"sc_female"`
	STMale      int `json:"st_male"`
	STFemale    int `json:"st_female"`
	OBCMale     int `json:"obc_male"`
	OBCFemale   int `json:"obc_female"`
	TotalMale   int `json:"total_male"`
	TotalFemale int `json:"total_female"`
	Total       int `json:"total"`
}

type DesignationDemo struct {
	Designation string    `json:"designation"`
	Stats       DemoStats `json:"stats"`
}
type ZoneDemo struct {
	Zone         string            `json:"zone"`
	Designations []DesignationDemo `json:"designations"`
}
type OverallDemo struct {
	Designations []DesignationDemo `json:"designations"`
}

var (
	EMP          = map[string]Emp{}
	SCH          = map[string]School{}
	DEMO_SCHOOLS []SchoolDemo // not used in UI v11.1 but kept for parity
	DEMO_ZONES   []ZoneDemo
	DEMO_OVERALL OverallDemo
)

var (
	rxBOM     = regexp.MustCompile("\uFEFF")
	rxNBSP    = regexp.MustCompile("\u00A0")
	rxDigits  = regexp.MustCompile(`(\d{5,})`)
	rxLastNum = regexp.MustCompile(`(\d{5,})\D*$`)
)

func norm(s string) string {
	s = rxBOM.ReplaceAllString(s, "")
	s = rxNBSP.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}
func stripDot0(s string) string {
	s = norm(s)
	if strings.HasSuffix(s, ".0") {
		return s[:len(s)-2]
	}
	return s
}
func digitsOnly(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			out = append(out, r)
		}
	}
	return string(out)
}
func normalizeEmpID(s string) string {
	s = stripDot0(s)
	s = digitsOnly(s)
	return s
}
func digitsOnlyKey(s string) string {
	s = norm(s)
	// Prefer the last number run in the string (common in "Name-1234567")
	if m := rxLastNum.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	// Fallback: any 5+ digit run
	if m := rxDigits.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	return ""
}
func atoiSafe(s string) int {
	s = stripDot0(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}
func readCSV(path string) (rows [][]string, header []string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	r := csv.NewReader(bufio.NewReader(f))
	r.FieldsPerRecord = -1
	header, err = r.Read()
	if err != nil {
		log.Fatalf("header %s: %v", path, err)
	}
	for i := range header {
		header[i] = norm(header[i])
	}
	for {
		rec, e := r.Read()
		if e != nil {
			break
		}
		for i := range rec {
			rec[i] = norm(rec[i])
		}
		rows = append(rows, rec)
	}
	return rows, header
}
func idxMap(h []string) map[string]int {
	m := map[string]int{}
	for i, c := range h {
		// Normalize invisible BOM or non-breaking spaces
		c = strings.ReplaceAll(c, "\u00A0", " ") // NBSP fix
		c = strings.ReplaceAll(c, "\uFEFF", "")  // BOM fix
		c = strings.Map(func(r rune) rune {
			if r == 160 || unicode.IsSpace(r) {
				return ' '
			}
			return r
		}, c)
		k := strings.ToLower(strings.TrimSpace(c))
		if _, ok := m[k]; !ok {
			m[k] = i
		}
	}
	return m
}

func get(rec []string, m map[string]int, name string, aliases ...string) string {
	names := append([]string{name}, aliases...)
	for _, nm := range names {
		i, ok := m[strings.ToLower(strings.TrimSpace(nm))]
		if ok && i >= 0 && i < len(rec) {
			return norm(rec[i])
		}
	}
	return ""
}

// parse dates like 19/Feb/1969, 19-Feb-1969, 19/02/1969, 19/February/1969
func parseDMYFlexible(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty")
	}
	// Normalize month to Title case
	parts := strings.Split(s, "/")
	if len(parts) == 3 && len(parts[1]) > 2 && !regexp.MustCompile(`^\d+$`).MatchString(parts[1]) {
		parts[1] = strings.ToUpper(parts[1][:1]) + strings.ToLower(parts[1][1:])
		s = strings.Join(parts, "/")
	}
	formats := []string{
		"02/Jan/2006", "2/Jan/2006", "02-Jan-2006", "2-Jan-2006",
		"02/01/2006", "2/01/2006", "02/January/2006", "2/January/2006",
	}
	var d time.Time
	var err error
	for _, f := range formats {
		d, err = time.Parse(f, s)
		if err == nil {
			return d, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparsed: %s", s)
}

func ageFromDOB(dob string) int {
	d, err := parseDMYFlexible(dob)
	if err != nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - d.Year()
	if now.Month() < d.Month() || (now.Month() == d.Month() && now.Day() < d.Day()) {
		age--
	}
	if age < 0 {
		return 0
	}
	return age
}

/* ===================== Email ===================== */
func sendEmail(to, subject, body string) error {
	if !LiveMode {
		fmt.Println("🧪 [Dry Run] To:", to, "| Sub:", subject)
		return nil
	}
	from := EmailFrom
	pass := EmailPass
	if from == "" || pass == "" {
		return fmt.Errorf("EMAIL_FROM / EMAIL_PASS not set")
	}
	msg := "From: " + EmailName + " <" + from + ">\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" + body

	auth := smtp.PlainAuth("", from, pass, "smtp.gmail.com")
	tlsconfig := &tls.Config{InsecureSkipVerify: true, ServerName: "smtp.gmail.com"}

	conn, err := tls.Dial("tcp", "smtp.gmail.com:465", tlsconfig)
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, "smtp.gmail.com")
	if err != nil {
		return err
	}
	if err = c.Auth(auth); err != nil {
		return err
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write([]byte(msg)); err != nil {
		return err
	}
	_ = w.Close()
	_ = c.Quit()
	return nil
}

func buildBirthdayEmail(name, gender string) (string, string) {
	sal := "Dear Sir/Madam"
	low := strings.ToLower(gender)
	if strings.HasPrefix(low, "m") {
		sal = "Dear Sir " + strings.Title(strings.ToLower(name))
	}
	if strings.HasPrefix(low, "f") {
		sal = "Dear Madam " + strings.Title(strings.ToLower(name))
	}
	sub := "🎉 Warm Birthday Wishes from HQ Team IT Education"
	body := fmt.Sprintf(`<p>%s,</p><p>Wishing you a very Happy Birthday! May your year be filled with health, happiness, and continued success.</p><p>Warm regards,<br><b>HQ Team IT Education</b></p>`, sal)
	return sub, body
}
func buildAnniversaryEmail(name, gender string, yrs int) (string, string) {
	sal := "Dear Sir/Madam"
	low := strings.ToLower(gender)
	if strings.HasPrefix(low, "m") {
		sal = "Dear Sir " + strings.Title(strings.ToLower(name))
	}
	if strings.HasPrefix(low, "f") {
		sal = "Dear Madam " + strings.Title(strings.ToLower(name))
	}
	sub := "🏅 Congratulations on your Work Anniversary!"
	body := fmt.Sprintf(`<p>%s,</p><p>Congratulations on completing <b>%d years</b> of dedicated service! Your work makes a real difference.</p><p>Warm regards,<br><b>HQ Team IT Education</b></p>`, sal, yrs)
	return sub, body
}

/* ===================== HTTP Handlers ===================== */

func handleToggleLive(w http.ResponseWriter, r *http.Request) {
	on := r.URL.Query().Get("on")
	if on == "1" {
		LiveMode = true
		fmt.Println("🚀 Live Mode ENABLED — real emails will be sent.")
	} else {
		LiveMode = false
		fmt.Println("🧪 Live Mode DISABLED — dry run only.")
	}
	fmt.Fprintf(w, "Mode: %v", LiveMode)
}

func handleSendBirthdays(w http.ResponseWriter, r *http.Request) {
	today := time.Now()
	count := 0
	validDOB := 0
	for _, rec := range bRows {
		dob := get(rec, bm, "Date of Birth", "DOB", "Birth Date", "D.O.B", "Date-of-Birth")
		empID := normalizeEmpID(get(rec, bm, "Employee ID", "Emp ID"))
		if empID == "" || dob == "" {
			continue
		}
		email := emailByEmpID[empID]
		name := nameByEmpID[empID]
		gender := genderByEmpID[empID]
		if email == "" {
			continue
		}
		d, err := parseDMYFlexible(dob)
		if err != nil {
			continue
		}
		validDOB++
		if d.Day() == today.Day() && d.Month() == today.Month() {
			sub, body := buildBirthdayEmail(name, gender)
			if err := sendEmail(email, sub, body); err == nil {
				count++
			}
		}
	}
	fmt.Fprintf(w, "🎂 Birthday emails sent to %d employees. (valid DOB parsed: %d)", count, validDOB)
}

func handleSendAnniversaries(w http.ResponseWriter, r *http.Request) {
	today := time.Now()
	count := 0
	validDOJ := 0
	for _, rec := range sRows {
		doj := get(rec, sm, "Date of Joining")
		empID := normalizeEmpID(get(rec, sm, "Employee ID", "Emp ID"))
		if empID == "" || doj == "" {
			continue
		}
		email := emailByEmpID[empID]
		name := nameByEmpID[empID]
		gender := genderByEmpID[empID]
		if email == "" {
			continue
		}
		d, err := parseDMYFlexible(doj)
		if err != nil {
			continue
		}
		validDOJ++
		if d.Day() == today.Day() && d.Month() == today.Month() {
			yrs := today.Year() - d.Year()
			sub, body := buildAnniversaryEmail(name, gender, yrs)
			if err := sendEmail(email, sub, body); err == nil {
				count++
			}
		}
	}
	fmt.Fprintf(w, "🏅 Anniversary emails sent to %d employees. (valid DOJ parsed: %d)", count, validDOJ)
}

/* ===================== Main ===================== */

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	var bh, sh, dh []string
	bRows, bh = readCSV(*basicCSV)
	sRows, sh = readCSV(*servCSV)
	dRows, dh = readCSV(*dbtCSV)
	bm, sm, dm = idxMap(bh), idxMap(sh), idxMap(dh)

	// Build lookup maps (emails, names, gender) from Basic
	for _, r := range bRows {
		id := normalizeEmpID(get(r, bm, "Employee ID", "Emp ID"))
		if id == "" {
			continue
		}
		if v := get(r, bm, "Employees_Email_ID"); v != "" {
			emailByEmpID[id] = v
		}
		if v := get(r, bm, "Name of the Employee", "Employee Name", "Name"); v != "" {
			nameByEmpID[id] = v
		}
		if v := get(r, bm, "Gender"); v != "" {
			genderByEmpID[id] = v
		}
		// Prefer Marital from Basic if present
		if v := get(r, bm, "Marital Status"); v != "" {
			maritalByEmpID[id] = v
		}
	}

	// DOJ + Category (from Services), with caching for DOJ
	log.Println("🔍 Preparing DOJ/Category maps...")
	dojByEmpID := map[string]string{}
	cacheFile := *cachePath
	svcInfo, _ := os.Stat(*servCSV)
	cacheInfo, err := os.Stat(cacheFile)
	needRebuild := err != nil || svcInfo.ModTime().After(cacheInfo.ModTime())
	if needRebuild || *forceRebuild {
		count := 0
		for _, r := range sRows {
			empID := normalizeEmpID(get(r, sm, "Employee ID", "Emp ID"))
			if empID == "" {
				continue
			}
			if dojRaw := get(r, sm, "Date of Joining"); dojRaw != "" {
				dojByEmpID[empID] = dojRaw
				count++
			}
			// Category enrichment from Services
			cat := strings.ToUpper(strings.TrimSpace(get(r, sm, "Selection Category", "")))
			if cat == "" {
				cat = strings.ToUpper(strings.TrimSpace(get(r, sm, "SelectionCategory", ""))) // fallback
			}
			if cat == "" {
				cat = strings.ToUpper(strings.TrimSpace(get(r, sm, "Applied Category", ""))) // secondary fallback
			}
			if cat == "" {
				cat = "UR" // default category if none found
			}
			catByEmpID[empID] = cat

			// If marital missing in Basic, try Services (rare)
			if mar := get(r, sm, "Marital Status", "Marital"); mar != "" {
				if _, ok := maritalByEmpID[empID]; !ok {
					maritalByEmpID[empID] = mar
				}
			}
		}

		log.Printf("Loaded SelectionCategory for %d employees", len(catByEmpID))

		cacheJSON, _ := json.Marshal(dojByEmpID)
		_ = os.MkdirAll(filepath.Dir(cacheFile), 0755)
		_ = os.WriteFile(cacheFile, cacheJSON, 0644)
		log.Printf("✅ Cached DOJ for %d employees → %s\n", count, cacheFile)
	} else {
		cacheBytes, err := os.ReadFile(cacheFile)
		if err != nil {
			log.Fatalf("❌ Error reading DOJ cache: %v", err)
		}
		_ = json.Unmarshal(cacheBytes, &dojByEmpID)
		for _, r := range sRows {
			empID := normalizeEmpID(get(r, sm, "Employee ID", "Emp ID"))
			if empID == "" {
				continue
			}
			if cat := strings.ToUpper(strings.TrimSpace(get(r, sm, "Selection Category", "Category"))); cat != "" {
				catByEmpID[empID] = cat
			}
			if mar := get(r, sm, "Marital Status", "Marital"); mar != "" {
				if _, ok := maritalByEmpID[empID]; !ok {
					maritalByEmpID[empID] = mar
				}
			}
		}
		log.Printf("⚡ Loaded DOJ cache (%d entries)", len(dojByEmpID))
	}

	zoneCounts := map[string]int{}
	desigCounts := map[string]int{}

	// Build EMP from Basic + enrich from Services
	joinCount := 0
	for _, r := range bRows {
		id := normalizeEmpID(get(r, bm, "Employee ID", "Emp ID"))
		if id == "" {
			continue
		}

		sname := get(r, bm, "School Name & ID", "School Name and ID", "School Name")
		sid := digitsOnlyKey(sname)

		zoneName := strings.ToUpper(strings.TrimSpace(get(r, bm, "Zone ID", "Zone Name", "Zone")))
		if zoneName == "" {
			zoneName = "UNKNOWN"
		}

		dob := get(r, bm, "Date of Birth", "DOB")

		e := Emp{
			ID:                id,
			Name:              nameByEmpID[id],
			Designation:       get(r, bm, "Designation"),
			DOB:               dob,
			Gender:            genderByEmpID[id],
			Zone:              zoneName,
			SchoolID:          sid,
			SchoolName:        sname,
			Status:            get(r, bm, "Status"),
			SelectionCategory: strings.ToUpper(strings.TrimSpace(catByEmpID[id])),
			MaritalStatus:     maritalByEmpID[id],
			Age:               ageFromDOB(dob),
			Mobile:            stripDot0(get(r, bm, "Mobile No.")),
			Email:             emailByEmpID[id],
			DOJ:               dojByEmpID[id],
		}
		if e.SelectionCategory != "" {
			joinCount++
		}
		EMP[id] = e

		if e.Zone == "" {
			e.Zone = "UNKNOWN"
		}
		zoneCounts[e.Zone]++
		d := e.Designation
		if d == "" {
			d = "UNKNOWN"
		}
		desigCounts[d]++
	}
	log.Printf("🔗 Services-enriched employees: %d (category attached)\n", joinCount)
	log.Printf("Loaded SelectionCategory for %d employees", len(catByEmpID))
	// Debug: Check category distribution
	catDebug := map[string]int{}
	for _, e := range EMP {
		c := e.SelectionCategory
		if c == "" {
			c = "EMPTY"
		}
		catDebug[c]++
	}
	log.Println("📊 Category Distribution:")
	for cat, count := range catDebug {
		log.Printf("  %s: %d employees", cat, count)
	}

	// Debug: Check first 5 employees
	log.Println("🔍 Sample Employees:")
	count := 0
	for _, e := range EMP {
		if count >= 5 {
			break
		}
		log.Printf("  ID:%s, Name:%s, Cat:%s, Gender:%s, Zone:%s, Desig:%s",
			e.ID, e.Name, e.SelectionCategory, e.Gender, e.Zone, e.Designation)
		count++
	}
	// Build SCH from DBT
	schoolByID := map[string]bool{}
	zoneSet := map[string]bool{}
	desigSet := map[string]bool{}

	for _, r := range dRows {
		sname := get(r, dm, "School Name & ID", "School Name")
		sid := digitsOnlyKey(sname)
		if sid == "" {
			continue
		}
		zoneName := strings.ToUpper(strings.TrimSpace(get(r, dm, "Zone ID", "Zone Name", "Zone")))
		if zoneName == "" {
			zoneName = "UNKNOWN"
		}
		SCH[sid] = School{
			ID:         sid,
			Name:       sname,
			Zone:       zoneName,
			SIName:     get(r, dm, "School Inspector's Name", "SI Name"),
			MaxTotal:   atoiSafe(get(r, dm, "Max Enrolment", "Total Enrolment (Last)", "Total Students")),
			MaxPresent: atoiSafe(get(r, dm, "Max Present", "Present Enrolment")),
			WithAcc:    atoiSafe(get(r, dm, "With Account")),
			WithoutAcc: atoiSafe(get(r, dm, "Without Account")),
			WithAad:    atoiSafe(get(r, dm, "With Aadhaar")),
			WithoutAad: atoiSafe(get(r, dm, "Without Aadhaar")),
			AadLinkAC:  atoiSafe(get(r, dm, "Aadhaar Linked Account")),
			NewMon:     atoiSafe(get(r, dm, "New Admission (This month)")),
			NewSess:    atoiSafe(get(r, dm, "New Admission (This session)")),
			DBTStud:    atoiSafe(get(r, dm, "DBT Received (Student)")),
			DBTPar:     atoiSafe(get(r, dm, "DBT Received (Parent)")),
			RecSP:      atoiSafe(get(r, dm, "Received By (Student + Parent)")),
		}
		schoolByID[sid] = true
		zoneSet[zoneName] = true
	}
	// Also collect sets from EMP in case DBT missing something
	for _, e := range EMP {
		if e.SchoolID != "" {
			schoolByID[e.SchoolID] = true
		}
		if e.Zone != "" {
			zoneSet[e.Zone] = true
		}
		if e.Designation != "" {
			desigSet[e.Designation] = true
		}
	}

	totalEmp := len(EMP)
	totalSchools := len(schoolByID)
	totalZones := len(zoneSet)
	totalDesigs := len(desigSet)

	type kv struct {
		K string
		V int
	}
	toSorted := func(m map[string]int) []kv {
		arr := make([]kv, 0, len(m))
		for k, v := range m {
			arr = append(arr, kv{k, v})
		}
		sort.Slice(arr, func(i, j int) bool {
			if arr[i].V == arr[j].V {
				return arr[i].K < arr[j].K
			}
			return arr[i].V > arr[j].V
		})
		return arr
	}
	take := func(a []kv, n int) []kv {
		if n > len(a) {
			n = len(a)
		}
		return a[:n]
	}
	zTop := take(toSorted(zoneCounts), 10)
	dTop := take(toSorted(desigCounts), 10)
	var ZLBL, ZVAL, DLBL, DVAL []string
	for _, p := range zTop {
		ZLBL = append(ZLBL, p.K)
		ZVAL = append(ZVAL, strconv.Itoa(p.V))
	}
	for _, p := range dTop {
		DLBL = append(DLBL, p.K)
		DVAL = append(DVAL, strconv.Itoa(p.V))
	}

	// ======= Build demographics (3 levels) =======
	log.Println("🔢 Building demographics...")

	// ---------- ZONE x DESIGNATION demographic map ----------
	zmap := map[string]map[string]*DemoStats{} // zone -> desig -> stats

	for _, e := range EMP {
		if e.Zone == "" || e.Designation == "" {
			continue
		}
		zone := strings.ToUpper(strings.TrimSpace(e.Zone))
		desig := strings.TrimSpace(e.Designation)
		cat := strings.ToUpper(strings.TrimSpace(e.SelectionCategory))
		g := strings.ToUpper(strings.TrimSpace(e.Gender))
		isMale := g == "MALE" || g == "M"
		isFemale := g == "FEMALE" || g == "F"

		if zmap[zone] == nil {
			zmap[zone] = map[string]*DemoStats{}
		}
		if zmap[zone][desig] == nil {
			zmap[zone][desig] = &DemoStats{}
		}
		stats := zmap[zone][desig]

		switch cat {
		case "UR", "GENERAL", "GEN":
			if isMale {
				stats.GenMale++
			}
			if isFemale {
				stats.GenFemale++
			}
		case "SC":
			if isMale {
				stats.SCMale++
			}
			if isFemale {
				stats.SCFemale++
			}
		case "ST":
			if isMale {
				stats.STMale++
			}
			if isFemale {
				stats.STFemale++
			}
		case "OBC":
			if isMale {
				stats.OBCMale++
			}
			if isFemale {
				stats.OBCFemale++
			}
		}

		if isMale {
			stats.TotalMale++
		}
		if isFemale {
			stats.TotalFemale++
		}
		stats.Total++ // ADD THIS LINE - it's missing in v11.1!
	}
	// ---------- Build DEMO_ZONES (slice) from zmap ----------
	DEMO_ZONES = nil
	for zone, dm := range zmap {
		var designations []DesignationDemo
		for dname, st := range dm {
			designations = append(designations, DesignationDemo{
				Designation: dname,
				Stats:       *st, // copy value (your DesignationDemo has `Stats DemoStats`)
			})
		}
		sort.Slice(designations, func(i, j int) bool { return designations[i].Designation < designations[j].Designation })
		DEMO_ZONES = append(DEMO_ZONES, ZoneDemo{
			Zone:         zone,
			Designations: designations,
		})
	}
	sort.Slice(DEMO_ZONES, func(i, j int) bool { return DEMO_ZONES[i].Zone < DEMO_ZONES[j].Zone })

	// ---------- Build DEMO_OVERALL (designation totals across all zones) ----------
	agg := map[string]*DemoStats{} // desig -> totals
	for _, z := range DEMO_ZONES {
		for _, d := range z.Designations {
			o := agg[d.Designation]
			if o == nil {
				o = &DemoStats{}
				agg[d.Designation] = o
			}
			s := d.Stats
			o.GenMale += s.GenMale
			o.GenFemale += s.GenFemale
			o.SCMale += s.SCMale
			o.SCFemale += s.SCFemale
			o.STMale += s.STMale
			o.STFemale += s.STFemale
			o.OBCMale += s.OBCMale
			o.OBCFemale += s.OBCFemale
			o.TotalMale += s.TotalMale
			o.TotalFemale += s.TotalFemale
			o.Total += s.Total
		}
	}

	DEMO_OVERALL = OverallDemo{}
	for dname, st := range agg {
		DEMO_OVERALL.Designations = append(DEMO_OVERALL.Designations, DesignationDemo{
			Designation: dname,
			Stats:       *st,
		})
	}
	sort.Slice(DEMO_OVERALL.Designations, func(i, j int) bool {
		return DEMO_OVERALL.Designations[i].Designation < DEMO_OVERALL.Designations[j].Designation
	})

	log.Printf("✅ Demographics built: zones=%d, overall desigs=%d",
		len(DEMO_ZONES), len(DEMO_OVERALL.Designations))
	// ======= Template injection =======
	tplBytes, err := os.ReadFile(*tplPath)
	if err != nil {
		log.Fatalf("read template: %v", err)
	}
	html := string(tplBytes)

	empJSON, _ := json.Marshal(EMP)
	schJSON, _ := json.Marshal(SCH)
	demoZonesJSON, _ := json.Marshal(DEMO_ZONES)
	demoOverallJSON, _ := json.Marshal(DEMO_OVERALL)
	staffJSON, _ := json.Marshal(buildStaff(SCH, EMP))

	zlblJSON, _ := json.Marshal(ZLBL)
	zvalJSON, _ := json.Marshal(ZVAL)
	dlblJSON, _ := json.Marshal(DLBL)
	dvalJSON, _ := json.Marshal(DVAL)

	// ======= Template injection =======
	tplBytes, err = os.ReadFile(*tplPath)
	if err != nil {
		log.Fatalf("read template: %v", err)
	}
	html = string(tplBytes)

	// Prepare category-wise totals for charts
	catTotals := map[string]int{"UR": 0, "SC": 0, "ST": 0, "OBC": 0}
	maleCount, femaleCount := 0, 0

	for _, e := range EMP {
		c := strings.ToUpper(strings.TrimSpace(e.SelectionCategory))
		switch {
		case strings.Contains(c, "SC"):
			catTotals["SC"]++
		case strings.Contains(c, "ST"):
			catTotals["ST"]++
		case strings.Contains(c, "OBC"):
			catTotals["OBC"]++
		default:
			catTotals["UR"]++
		}
		if strings.HasPrefix(strings.ToUpper(e.Gender), "M") {
			maleCount++
		} else if strings.HasPrefix(strings.ToUpper(e.Gender), "F") {
			femaleCount++
		}
	}

	catJSON, _ := json.Marshal(catTotals)
	genderJSON, _ := json.Marshal(map[string]int{"Male": maleCount, "Female": femaleCount})

	empJSON, _ = json.Marshal(EMP)
	schJSON, _ = json.Marshal(SCH)
	demoZonesJSON, _ = json.Marshal(DEMO_ZONES)
	demoOverallJSON, _ = json.Marshal(DEMO_OVERALL)
	staffJSON, _ = json.Marshal(buildStaff(SCH, EMP))

	zlblJSON, _ = json.Marshal(ZLBL)
	zvalJSON, _ = json.Marshal(ZVAL)
	dlblJSON, _ = json.Marshal(DLBL)
	dvalJSON, _ = json.Marshal(DVAL)

	html = strings.ReplaceAll(html, "{{EMP_JSON}}", string(empJSON))
	html = strings.ReplaceAll(html, "{{SCH_JSON}}", string(schJSON))
	html = strings.ReplaceAll(html, "{{STAFF_JSON}}", string(staffJSON))
	html = strings.ReplaceAll(html, "{{DEMO_SCHOOLS_JSON}}", "[]")
	html = strings.ReplaceAll(html, "{{DEMO_ZONES_JSON}}", string(demoZonesJSON))
	html = strings.ReplaceAll(html, "{{DEMO_OVERALL_JSON}}", string(demoOverallJSON))
	html = strings.ReplaceAll(html, "{{CATEGORY_WISE_JSON}}", string(catJSON))
	html = strings.ReplaceAll(html, "{{GENDER_WISE_JSON}}", string(genderJSON))
	html = strings.ReplaceAll(html, "{{ZLBL}}", string(zlblJSON))
	html = strings.ReplaceAll(html, "{{ZVAL}}", string(zvalJSON))
	html = strings.ReplaceAll(html, "{{DLBL}}", string(dlblJSON))
	html = strings.ReplaceAll(html, "{{DVAL}}", string(dvalJSON))
	html = strings.ReplaceAll(html, "{{TOTAL_EMP}}", strconv.Itoa(totalEmp))
	html = strings.ReplaceAll(html, "{{TOTAL_SCHOOLS}}", strconv.Itoa(totalSchools))
	html = strings.ReplaceAll(html, "{{TOTAL_ZONES}}", strconv.Itoa(totalZones))
	html = strings.ReplaceAll(html, "{{TOTAL_DESIGS}}", strconv.Itoa(totalDesigs))
	html = strings.ReplaceAll(html, "<title>MCD Dashboard</title>", "<title>"+*title+"</title>")

	_ = os.MkdirAll(filepath.Dir(*outHTML), 0755)
	if err := os.WriteFile(*outHTML, []byte(html), 0644); err != nil {
		log.Fatal(err)
	}
	log.Printf("OK → %s", *outHTML)
	log.Printf("Counts → EMP:%d, Schools:%d, Zones:%d, Desigs:%d", totalEmp, totalSchools, totalZones, totalDesigs)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, *outHTML) })
	mux.HandleFunc("/toggle-live", handleToggleLive)
	mux.HandleFunc("/send-birthdays", handleSendBirthdays)
	mux.HandleFunc("/send-anniversaries", handleSendAnniversaries)
	log.Printf("Server running on http://localhost%v", *listen)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		log.Fatal(err)
	}
} // 👈 close main() here

// Build STAFF summary from SCH + EMP
func buildStaff(SCH map[string]School, EMP map[string]Emp) []SchoolStaff {
	result := []SchoolStaff{}
	for sid, s := range SCH {
		actualT := 0
		hasP := false
		hasSE := false
		totalSt := 0
		for _, e := range EMP {
			if e.SchoolID == sid {
				totalSt++
				desLow := strings.ToLower(e.Designation)
				if strings.Contains(desLow, "teacher") {
					actualT++
				} else if strings.Contains(desLow, "principal") {
					hasP = true
				} else if strings.Contains(desLow, "special educator") {
					hasSE = true
				}
			}
		}
		neededT := int(math.Ceil(float64(s.MaxPresent) / 40.0))
		if neededT < 0 {
			neededT = 0
		}
		sv := actualT - neededT
		ratio := 0.0
		if actualT > 0 {
			ratio = float64(s.MaxPresent) / float64(actualT)
		}
		result = append(result, SchoolStaff{
			ID:             sid,
			Name:           s.Name,
			Zone:           s.Zone,
			NeededTeachers: neededT,
			ActualTeachers: actualT,
			SurplusVacancy: sv,
			HasPrincipal:   hasP,
			HasSpecialEdu:  hasSE,
			TotalStaff:     totalSt,
			Ratio:          ratio,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}
