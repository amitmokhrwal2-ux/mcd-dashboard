// mcd_dashboard_v7.go
// Complete dashboard builder + server (DBT merge, search, filters, birthdays & work-anniversary emails).
// Uses Employees_Email_ID ONLY from Basic.csv. Gmail App Password is hardcoded as placeholders below.
// Modern dark university-like UI via template (mcd_dashboard_template_v7.html).
//
// Build & run:
//   go run mcd_dashboard_v7.go \
//     -basic ./Basic.csv \
//     -services ./Services.csv \
//     -dbt ./Dashboard_Summary_202509.csv \
//     -template ./mcd_dashboard_template_v7.html \
//     -out ./out/mcd_dashboard_202509.html \
//     -title "MCD Dashboard"
//
// Then open http://localhost:8080
//
// -----------------------------------------------------------------------------

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
)

/* ===================== Flags & Globals ===================== */

var (
	basicCSV     = flag.String("basic", "Basic.csv", "Basic CSV path")
	servCSV      = flag.String("services", "Services.csv", "Services CSV path")
	dbtCSV       = flag.String("dbt", "Dashboard_Summary_202509.csv", "DBT CSV path")
	tplPath      = flag.String("template", "mcd_dashboard_template_v7.html", "HTML template path")
	outHTML      = flag.String("out", "./out/mcd_dashboard_202509.html", "Output HTML path")
	title        = flag.String("title", "MCD Dashboard", "Document title")
	listen       = flag.String("listen", ":8080", "HTTP listen address")
	cachePath    = flag.String("cache", "./out/doj_cache.json", "DOJ cache path (optional)")
	forceRebuild = flag.Bool("rebuild", false, "Force rebuild cache") // <-- ADD THIS LINE
	// ======= Gmail creds (placeholders; fill in before sending) =======
	EmailFrom = "amitmokhrwal2@gmail.com" // <-- put your Gmail here
	EmailPass = "uxlorpczyrdkzwfj"        // <-- put your 16-char app password here
	EmailName = "HQ Team IT Education"

	// Real send (true) vs dry-run (false). You earlier asked real send; default true.
	LiveMode = true
)

// CSV state available to handlers
var (
	bRows, sRows, dRows [][]string
	bm, sm, dm          map[string]int
)

// Lookups (from Basic.csv only)
var (
	emailByEmpID  = map[string]string{}
	nameByEmpID   = map[string]string{}
	genderByEmpID = map[string]string{}
)
var (
	// ... existing vars
	zoneMap map[string]string // <-- ADD THIS: numeric -> readable name
)

// Data we inject to HTML (lightweight shapes for UI JS)
type Emp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Designation string `json:"designation"`
	Gender      string `json:"gender"`
	Zone        string `json:"zone"`
	SchoolID    string `json:"school_id"`
	SchoolName  string `json:"school_name"`
	Status      string `json:"status"`

	// --- Personal ---
	FatherName    string `json:"father_name"`
	MotherName    string `json:"mother_name"`
	MaritalStatus string `json:"marital_status"`
	SpouseName    string `json:"spouse_name"`
	Religion      string `json:"religion"`
	PhysicalCat   string `json:"physical_category"`
	HomeTown      string `json:"home_town"`
	Accommodation string `json:"accommodation"`

	// --- Contact ---
	Mobile         string `json:"mobile"`
	Email          string `json:"email"`
	EmergencyName  string `json:"emergency_name"`
	EmergencyNo    string `json:"emergency_no"`
	CorrespondAddr string `json:"correspond_address"`
	PermanentAddr  string `json:"permanent_address"`

	// --- Education ---
	EduQual  string `json:"edu_qual"`
	ProfQual string `json:"prof_qual"`

	// --- Service ---
	DOJ           string `json:"doj"`
	DateOfAppt    string `json:"date_of_appointment"`
	AppType       string `json:"appointment_type"`
	PromotionDate string `json:"promotion_date"`
	TransferDate  string `json:"transfer_date"`
	Deputed       string `json:"deputed"`
	LastTraining  string `json:"last_training"`

	// --- Pay & Bank ---
	PayLevel    string `json:"pay_level"`
	BasicPay    string `json:"basic_pay"`
	PAN         string `json:"pan"`
	PRAN_GPF    string `json:"pran_gpf"`
	PensionType string `json:"pension_type"`

	// --- Misc ---
	Messenger    string `json:"messenger"`
	MessengerNo  string `json:"messenger_no"`
	ExtraDuty    string `json:"extra_duty"`
	MedicalIssue string `json:"medical_issue"`
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
	NeededTeachers int     `json:"needed_teachers"` // MaxPresent / 40
	ActualTeachers int     `json:"actual_teachers"` // EMP filter
	SurplusVacancy int     `json:"surplus_vacancy"` // Actual - Needed
	HasPrincipal   bool    `json:"has_principal"`
	HasSpecialEdu  bool    `json:"has_special_educator"`
	TotalStaff     int     `json:"total_staff"`
	Ratio          float64 `json:"ratio"` // Students / Teachers
}

var (
	EMP = map[string]Emp{}    // id -> Emp
	SCH = map[string]School{} // school id -> School
)

/* ===================== Utils ===================== */

var (
	rxBOM    = regexp.MustCompile("\uFEFF")
	rxNBSP   = regexp.MustCompile("\u00A0")
	rxDigits = regexp.MustCompile(`(\d{5,})`)
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
func digitsOnlyKey(s string) string {
	s = norm(s)
	if m := rxDigits.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	return rxDigits.FindString(s)
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

// robust dd/Mon/yyyy parsing (and variants)
func parseDayMonth(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty")
	}
	// Normalize month capitalization
	if len(s) > 3 {
		s = strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
	}
	formats := []string{"02/Jan/2006", "2/Jan/2006", "02-Jan-2006", "02/01/2006"}
	var d time.Time
	var err error
	for _, f := range formats {
		d, err = time.Parse(f, s)
		if err == nil {
			return d, nil
		}
	}
	return time.Time{}, err
}

/* ===================== Email ===================== */

func sendEmail(to, subject, body string) error {
	if !LiveMode {
		// Keep dry-run path available if you flip the toggle endpoint.
		fmt.Println("🧪 [Dry Run] To:", to, "| Sub:", subject)
		return nil
	}
	from := EmailFrom
	pass := EmailPass
	if from == "" || pass == "" {
		return fmt.Errorf("EMAIL_FROM / EMAIL_PASS not set in code")
	}
	msg := "From: " + EmailName + " <" + from + ">\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		body

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
	body := fmt.Sprintf(`
<p>%s,</p>
<p>Wishing you a very Happy Birthday! May your year be filled with health, happiness, and continued success.</p>
<p>Stay updated with resources and tools:<br>
👉 <a href="https://whatsapp.com/channel/0029Vb6hLZd1CYoIFek51P0V">Follow our official channel</a>
</p>
<p>Warm regards,<br><b>HQ Team IT Education</b><br><i>Empowering Teachers • Simplifying Systems</i></p>
`, sal)
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
	body := fmt.Sprintf(`
<p>%s,</p>
<p>Congratulations on completing <b>%d years</b> of dedicated service! Your work makes a real difference.</p>
<p>Stay updated with resources and tools:<br>
👉 <a href="https://whatsapp.com/channel/0029Vb6hLZd1CYoIFek51P0V">Follow our official channel</a>
</p>
<p>Warm regards,<br><b>HQ Team IT Education</b><br><i>Empowering Teachers • Simplifying Systems</i></p>
`, sal, yrs)
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

	for _, rec := range bRows {
		dob := get(rec, bm, "Date of Birth", "DOB", "date of birth", "dob", "Birth Date", "D.O.B", "Date-of-Birth")

		empID := stripDot0(get(rec, bm, "Employee ID", "Emp ID"))
		if dob == "" {
			log.Printf("⚠️ Missing DOB for EmpID=%s Name=%s", empID, nameByEmpID[empID])
		}

		if empID == "" || dob == "" {
			continue
		}
		email := emailByEmpID[empID] // ONLY from Basic.csv
		name := nameByEmpID[empID]
		gender := genderByEmpID[empID]
		if email == "" {
			continue
		}
		d, err := parseDayMonth(dob)
		if err != nil {
			continue
		}
		if d.Day() == today.Day() && d.Month() == today.Month() {
			log.Printf("🎂 Birthday match → %s (%s) DOB=%s", name, empID, dob)

			sub, body := buildBirthdayEmail(name, gender)
			if err := sendEmail(email, sub, body); err == nil {
				count++
			} else {
				log.Printf("send birthday → %s failed: %v", email, err)
			}
		}
	}
	fmt.Fprintf(w, "🎂 Birthday emails sent to %d employees.", count)
}

func handleSendAnniversaries(w http.ResponseWriter, r *http.Request) {
	today := time.Now()
	count := 0

	for _, rec := range sRows {
		doj := get(rec, sm, "Date of Joining")
		empID := stripDot0(get(rec, sm, "Employee ID", "Emp ID"))
		if empID == "" || doj == "" {
			continue
		}
		email := emailByEmpID[empID] // ONLY from Basic.csv
		name := nameByEmpID[empID]
		gender := genderByEmpID[empID]
		if email == "" {
			continue
		}
		d, err := parseDayMonth(doj)
		if err != nil {
			continue
		}
		if d.Day() == today.Day() && d.Month() == today.Month() {
			yrs := today.Year() - d.Year()
			sub, body := buildAnniversaryEmail(name, gender, yrs)
			if err := sendEmail(email, sub, body); err == nil {
				count++
			} else {
				log.Printf("send anniversary → %s failed: %v", email, err)
			}
		}
	}
	fmt.Fprintf(w, "🏅 Anniversary emails sent to %d employees.", count)
}

/* ===================== Main ===================== */

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// Read CSVs
	var bh, sh, dh []string
	bRows, bh = readCSV(*basicCSV)
	sRows, sh = readCSV(*servCSV)
	dRows, dh = readCSV(*dbtCSV)
	bm, sm, dm = idxMap(bh), idxMap(sh), idxMap(dh)
	// Build zone mapping from Basic.csv (unique numeric "Zone" -> "Zone ID")
	// ===================== DOJ CACHE OPTIMIZATION =====================
	// This block ensures Date of Joining (DOJ) data is built only once and cached.
	// On next runs, it loads instantly from cache.json (no 22,000x reload).

	log.Println("🔍 Checking DOJ cache...")

	dojByEmpID := map[string]string{}
	cacheFile := *cachePath

	// Check whether cache file exists and if rebuild is required
	svcInfo, _ := os.Stat(*servCSV)
	cacheInfo, err := os.Stat(cacheFile)
	needRebuild := err != nil || svcInfo.ModTime().After(cacheInfo.ModTime())

	if needRebuild || *forceRebuild {
		log.Println("🧱 Rebuilding DOJ map from Services.csv ...")
		count := 0
		for _, r := range sRows {
			empID := stripDot0(get(r, sm, "Employee ID", "Emp ID"))
			dojRaw := get(r, sm, "Date of Joining")
			if empID != "" && dojRaw != "" {
				dojByEmpID[empID] = dojRaw
				count++
			}
		}
		cacheJSON, _ := json.Marshal(dojByEmpID)
		os.WriteFile(cacheFile, cacheJSON, 0644)
		log.Printf("✅ Cached DOJ for %d employees → %s\n", count, cacheFile)
	} else {
		// Load from cache directly
		cacheBytes, err := os.ReadFile(cacheFile)
		if err != nil {
			log.Fatalf("❌ Error reading DOJ cache: %v", err)
		}
		json.Unmarshal(cacheBytes, &dojByEmpID)
		log.Printf("⚡ Loaded DOJ cache from %s (%d entries)\n", cacheFile, len(dojByEmpID))
	}

	zoneMap = map[string]string{}
	for _, r := range bRows {
		// ----- Zone detection (handles both Zone ID and numeric Zone) -----
		// ----- Zone detection (DBT CSV) -----
		zoneName := get(r, dm, "Zone ID", "Zone Name", "Zone")
		if zoneName == "" {
			rawZone := get(r, dm, "Zone", "Zone Name")
			if mapped, ok := zoneMap[rawZone]; ok {
				zoneName = mapped
			} else {
				zoneName = rawZone
			}
		}
		if zoneName == "" {
			zoneName = "Unknown"
		}

	}
	log.Printf("Loaded %d unique zone mappings from Basic.csv", len(zoneMap)) // Optional: for debug
	// Build Basic lookups (email/name/gender) and EMP map for UI
	zoneCounts := map[string]int{}
	desigCounts := map[string]int{}
	for _, r := range bRows {
		id := stripDot0(get(r, bm, "Employee ID", "Emp ID"))
		if id == "" {
			continue
		}
		email := get(r, bm, "Employees_Email_ID")
		if email != "" {
			emailByEmpID[id] = email
		}
		nameByEmpID[id] = get(r, bm, "Name of the Employee", "Employee Name", "Name")
		genderByEmpID[id] = get(r, bm, "Gender")

		sname := get(r, bm, "School Name & ID", "School Name and ID", "School Name")
		sid := digitsOnlyKey(sname)
		// Build DOJ lookup from Services.csv (empID -> DOJ string)
		// Build DOJ map (as before)
		// Build DOJ map (as before)

		log.Printf("Loaded DOJ for %d employees from Services.csv", len(dojByEmpID)) // Optional log
		// Inside the loop:
		// Zone column fix (Basic.csv now has "Zone ID" like Central, South, etc.)
		zoneName := get(r, bm, "Zone ID", "Zone Name", "Zone")
		if zoneName == "" {
			zoneName = "Unknown"
		}

		e := Emp{
			ID:          id,
			Name:        nameByEmpID[id],
			Designation: get(r, bm, "Designation"),
			Gender:      genderByEmpID[id],
			Zone:        zoneName,
			SchoolID:    sid,
			SchoolName:  sname,
			Status:      get(r, bm, "Status"),

			// Personal
			FatherName:    get(r, bm, "Father's Name"),
			MotherName:    get(r, bm, "Mother's Name"),
			MaritalStatus: get(r, bm, "Marital Status"),
			SpouseName:    get(r, bm, "Spouse Name"),
			Religion:      get(r, bm, "Religion"),
			PhysicalCat:   get(r, bm, "Physical Category"),
			HomeTown:      get(r, bm, "Home Town"),
			Accommodation: get(r, bm, "Accomodation Type"),

			// Contact
			Mobile:         stripDot0(get(r, bm, "Mobile No.")),
			Email:          get(r, bm, "Employees_Email_ID"),
			EmergencyName:  get(r, bm, "Emergency Name & Relation"),
			EmergencyNo:    stripDot0(get(r, bm, "Emergency No.")),
			CorrespondAddr: get(r, bm, "Correspondance_Address"),
			PermanentAddr:  get(r, bm, "Permanent_Address"),

			// Education
			EduQual:  get(r, bm, "Edn. Qual. @ Present"),
			ProfQual: get(r, bm, "Prof. Qual. @ Present"),

			// Pay & Bank
			PayLevel:    get(r, bm, "Pay Level"),
			BasicPay:    get(r, bm, "Basic Pay"),
			PAN:         get(r, bm, "PAN"),
			PRAN_GPF:    get(r, bm, "PRAN / GPF No."),
			PensionType: get(r, bm, "Pension Type"),

			// Misc
			Messenger:    get(r, bm, "Messenger"),
			MessengerNo:  get(r, bm, "Messenger No."),
			ExtraDuty:    get(r, bm, "Performing Extra Duty"),
			MedicalIssue: get(r, bm, "Medical Issue"),

			// From Services.csv (merge by EmpID)
			DOJ: dojByEmpID[id],
		}

		EMP[id] = e

		// ... rest same (zoneCounts will now use mapped names)

		z := e.Zone
		if z == "" {
			z = "Unknown"
		}
		zoneCounts[z]++
		d := e.Designation
		if d == "" {
			d = "Unknown"
		}
		desigCounts[d]++
	}

	// Build SCH (DBT merge by numeric tail from "School Name & ID")
	for _, r := range dRows {
		sname := get(r, dm, "School Name & ID", "School Name")
		sid := digitsOnlyKey(sname)
		if sid == "" {
			continue
		}
		// Inside the loop:
		zoneName := get(r, dm, "Zone ID", "Zone Name", "Zone")
		if zoneName == "" {
			zoneName = "Unknown"
		}

		SCH[sid] = School{
			ID:         sid,
			Name:       sname,
			Zone:       zoneName, // <-- USE MAPPED NAME
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
	}
	var STAFF []SchoolStaff // Array for JSON
	for sid, s := range SCH {
		// Teachers count from EMP
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
				} else if strings.Contains(desLow, "special educator") { // Adjust keyword if needed
					hasSE = true
				}
			}
		}
		neededT := int(math.Ceil(float64(s.MaxPresent) / 40.0))
		sv := actualT - neededT
		ratio := 0.0
		if actualT > 0 {
			ratio = float64(s.MaxPresent) / float64(actualT)
		}
		STAFF = append(STAFF, SchoolStaff{
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
	sort.Slice(STAFF, func(i, j int) bool { return STAFF[i].Name < STAFF[j].Name }) // Sort by name
	staffJSON, _ := json.Marshal(STAFF)
	// Aggregate counters
	totalEmp := len(EMP)
	sIDset := map[string]bool{}
	zoneSet := map[string]bool{}
	desigSet := map[string]bool{}
	for _, e := range EMP {
		if e.SchoolID != "" {
			sIDset[e.SchoolID] = true
		}
		if e.Zone != "" {
			zoneSet[e.Zone] = true
		}
		if e.Designation != "" {
			desigSet[e.Designation] = true
		}
	}
	totalSchools := len(sIDset)
	totalZones := len(zoneSet)
	totalDesigs := len(desigSet)

	// Top-N arrays for charts
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
		ZLBL, ZVAL = append(ZLBL, p.K), append(ZVAL, strconv.Itoa(p.V))
	}
	for _, p := range dTop {
		DLBL, DVAL = append(DLBL, p.K), append(DVAL, strconv.Itoa(p.V))
	}
	zlblJSON, _ := json.Marshal(ZLBL)
	zvalJSON, _ := json.Marshal(ZVAL)
	dlblJSON, _ := json.Marshal(DLBL)
	dvalJSON, _ := json.Marshal(DVAL)

	// Marshal JSON
	empJSON, _ := json.Marshal(EMP)
	schJSON, _ := json.Marshal(SCH)
	staffJSONStr := string(staffJSON)
	// Read template
	tplBytes, err := os.ReadFile(*tplPath)
	if err != nil {
		log.Fatalf("read template: %v", err)
	}
	html := string(tplBytes)

	// Title
	html = strings.Replace(html, "<title>MCD Dashboard</title>",
		fmt.Sprintf("<title>%s</title>", *title), 1)

	// Inject JSON (works whether placeholders exist or not)
	if strings.Contains(html, "{{EMP_JSON}}") {
		html = strings.ReplaceAll(html, "{{EMP_JSON}}", string(empJSON))
	} else {
		html = strings.Replace(html, "</head>", "<script>const EMP="+string(empJSON)+";</script></head>", 1)
	}
	if strings.Contains(html, "{{SCH_JSON}}") {
		html = strings.ReplaceAll(html, "{{SCH_JSON}}", string(schJSON))
	} else {
		html = strings.Replace(html, "</head>", "<script>const SCH="+string(schJSON)+";</script></head>", 1)
	}
	if strings.Contains(html, "{{STAFF_JSON}}") {
		html = strings.ReplaceAll(html, "{{STAFF_JSON}}", staffJSONStr) // <-- YOUR LINE HERE
	} else {
		// Fallback: Inject as script if placeholder missing
		html = strings.Replace(html, "</head>", "<script>const STAFF="+staffJSONStr+";</script></head>", 1)
	}

	// Header counters (existing)
	html = strings.ReplaceAll(html, "{{TOTAL_EMP}}", strconv.Itoa(totalEmp))
	// ... (other totals)

	// Write output (existing)
	if err := os.WriteFile(*outHTML, []byte(html), 0o644); err != nil {
		log.Fatal(err)
	}
	// Inject header counters
	html = strings.ReplaceAll(html, "{{TOTAL_EMP}}", strconv.Itoa(totalEmp))
	html = strings.ReplaceAll(html, "{{TOTAL_SCHOOLS}}", strconv.Itoa(totalSchools))
	html = strings.ReplaceAll(html, "{{TOTAL_ZONES}}", strconv.Itoa(totalZones))
	html = strings.ReplaceAll(html, "{{TOTAL_DESIGS}}", strconv.Itoa(totalDesigs))

	// Inject charts series (optional; used by template if present)
	html = strings.ReplaceAll(html, "{{ZLBL}}", string(zlblJSON))
	html = strings.ReplaceAll(html, "{{ZVAL}}", string(zvalJSON))
	html = strings.ReplaceAll(html, "{{DLBL}}", string(dlblJSON))
	html = strings.ReplaceAll(html, "{{DVAL}}", string(dvalJSON))

	// Write output file
	if err := os.MkdirAll(filepath.Dir(*outHTML), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*outHTML, []byte(html), 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("OK → %s", *outHTML)

	// Minimal server to handle email endpoints and static serving of HTML
	http.HandleFunc("/toggle_live", handleToggleLive)
	http.HandleFunc("/send_birthdays", handleSendBirthdays)
	http.HandleFunc("/send_anniversaries", handleSendAnniversaries)

	// Serve the generated HTML at /
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, *outHTML)
	})

	log.Printf("Server running on http://localhost%s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
