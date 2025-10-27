// mcd_dashboard_v16_fixed.go
// Live dashboard server (localhost:8080) with:
// - Charts on top (2x2 grid)
// - Advanced employee filter (incl. Marital)
// - Employee & School lookup (by ID/Name)
// - Dynamic Category & Religion tables with their own filters (zone/designation + category/religion)
// - Wide, horizontally scrollable tables
// - Proper normalization for category/religion keys so column counts match totals
// - Fixed issue with zero values in category and religion tables
//
// Run (Windows PowerShell):
//
//	go run mcd_dashboard_v16_fixed.go ^
//	  -basic ./Basic.csv ^
//	  -services ./Services.csv ^
//	  -dbt ./Dashboard_Summary_202509.csv ^
//	  -title "MCD Dashboard"
//
// Open: http://localhost:8080
package main

import (
	"bufio"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"html"
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

// ---- Flags ----
var (
	basicCSV  = flag.String("basic", "Basic.csv", "Basic CSV path")
	servCSV   = flag.String("services", "Services.csv", "Services CSV path")
	dbtCSV    = flag.String("dbt", "Dashboard_Summary_202509.csv", "DBT CSV path")
	title     = flag.String("title", "MCD Dashboard", "Page title")
	listen    = flag.String("listen", ":8080", "HTTP listen address")
	cachePath = flag.String("cache", "./out/doj_cache.json", "DOJ cache path (optional)")

	EmailFrom = "amitmokhrwal2@gmail.com" // optional for mailers
	EmailPass = "uxlorpczyrdkzwfj"
	EmailName = "HQ Team IT Education"
	LiveMode  = false
)

// ---- Models ----
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
	Mobile            string `json:"mobile"`
	Email             string `json:"email"`
	DOJ               string `json:"doj"`
	FatherName        string `json:"father_name"`
	MotherName        string `json:"mother_name"`
	SpouseName        string `json:"spouse_name"`
	Correspondence    string `json:"correspondence"`
	Permanent         string `json:"permanent"`
	HomeTown          string `json:"home_town"`
	Religion          string `json:"religion"`
	AppointmentDate   string `json:"appointment_date"`
	PromotionDate     string `json:"promotion_date"`
	TransferDate      string `json:"transfer_date"`
}

type School struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Zone                string `json:"zone"`
	SIName              string `json:"si_name"`
	TotalEnrolment      int    `json:"total_enrolment"`
	MaxEnrolment        int    `json:"max_enrolment"`
	MaxEnrolmentDate    string `json:"max_enrolment_date"`
	MaxPresent          int    `json:"max_present"`
	MaxPresentDate      string `json:"max_present_date"`
	WithAccount         int    `json:"with_account"`
	WithoutAccount      int    `json:"without_account"`
	WithAadhaar         int    `json:"with_aadhaar"`
	WithoutAadhaar      int    `json:"without_aadhaar"`
	AadhaarLinkedAcc    int    `json:"aadhaar_linked_acc"`
	NewAdmissionMonth   int    `json:"new_admission_month"`
	NewAdmissionSession int    `json:"new_admission_session"`
	DBTStudent          int    `json:"dbt_student"`
	DBTParent           int    `json:"dbt_parent"`
	DBTTotal            int    `json:"dbt_total"`
}

type Stat struct {
	Male   int `json:"male"`
	Female int `json:"female"`
}

type DemoStats struct {
	TotalMale     int              `json:"total_male"`
	TotalFemale   int              `json:"total_female"`
	Total         int              `json:"total"`
	CatStats      map[string]*Stat `json:"cat_stats,omitempty"`
	ReligionStats map[string]*Stat `json:"religion_stats,omitempty"`
}

type DesignationDemo struct {
	Designation string    `json:"designation"`
	Stats       DemoStats `json:"stats"`
}

type ZoneDemo struct {
	Zone         string            `json:"zone"`
	Designations []DesignationDemo `json:"designations"`
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

// ---- Globals ----
var (
	EMP = map[string]Emp{}
	SCH = map[string]School{}

	bRows, sRows, dRows [][]string
	bm, sm, dm          map[string]int

	emailByEmpID   = map[string]string{}
	nameByEmpID    = map[string]string{}
	genderByEmpID  = map[string]string{}
	catByEmpID     = map[string]string{}
	maritalByEmpID = map[string]string{}

	ZLBL, ZVAL, DLBL, DVAL []string

	DEMO_ZONES []ZoneDemo

	CATEGORY_WISE = map[string]int{}
	GENDER_WISE   = map[string]int{}
)

// ---- Utils ----
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
func normalizeEmpID(s string) string { return digitsOnly(stripDot0(s)) }
func digitsOnlyKey(s string) string {
	s = norm(s)
	if m := rxLastNum.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
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
		c = strings.ReplaceAll(c, "\u00A0", " ")
		c = strings.ReplaceAll(c, "\uFEFF", "")
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
func parseDMYFlexible(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty")
	}
	parts := strings.Split(s, "/")
	if len(parts) == 3 && len(parts[1]) > 2 && !regexp.MustCompile(`^\d+$`).MatchString(parts[1]) {
		parts[1] = strings.ToUpper(strings.ToUpper(parts[1][:1]) + strings.ToLower(parts[1][1:]))
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

// ---- Category Canonicalization ----
func canonicalCategory(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	if s == "" {
		return "UNKNOWN"
	}
	log.Printf("Category: Raw=%q, Normalized=%q", raw, s)
	return s
}

// ---- Religion Canonicalization ----
func canonicalReligion(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	if s == "" {
		return "UNKNOWN"
	}
	log.Printf("Religion: Raw=%q, Normalized=%q", raw, s)
	return s
}

// ---- Email helpers (unchanged) ----
func sendEmail(to, subject, body string) error {
	if !LiveMode {
		log.Printf("DRY SEND → %s | %s", to, subject)
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

// ---- Build data ----
func buildAll() {
	// Read CSVs
	var bh, sh, dh []string
	bRows, bh = readCSV(*basicCSV)
	sRows, sh = readCSV(*servCSV)
	dRows, dh = readCSV(*dbtCSV)
	bm, sm, dm = idxMap(bh), idxMap(sh), idxMap(dh)

	// Lookup maps
	log.Println("📋 Building lookup maps from Basic.csv...")
	for _, r := range bRows {
		id := normalizeEmpID(get(r, bm, "Employee ID", "Emp ID"))
		if id == "" {
			continue
		}
		if v := get(r, bm, "Employees_Email_ID", "Email", "Email ID"); v != "" {
			emailByEmpID[id] = v
		}
		if v := get(r, bm, "Name of the Employee", "Employee Name", "Name"); v != "" {
			nameByEmpID[id] = v
		}
		if v := get(r, bm, "Gender"); v != "" {
			genderByEmpID[id] = strings.ToUpper(v)
		}
		if v := get(r, bm, "Marital Status", "Marital"); v != "" {
			maritalByEmpID[id] = v
		}
	}

	// Services: DOJ + Category + dates
	log.Println("📁 Preparing DOJ/Category maps...")
	dojByEmpID := map[string]string{}
	for _, r := range sRows {
		empID := normalizeEmpID(get(r, sm, "Employee ID", "Emp ID"))
		if empID == "" {
			continue
		}
		if dojRaw := get(r, sm, "Date of Joining", "DOJ"); dojRaw != "" {
			dojByEmpID[empID] = dojRaw
		}
		rawCat := get(r, sm, "Selection Category", "SelectionCategory", "Applied Category", "Category")
		cat := canonicalCategory(rawCat)
		catByEmpID[empID] = cat
		if mar := get(r, sm, "Marital Status", "Marital"); mar != "" {
			if _, ok := maritalByEmpID[empID]; !ok {
				maritalByEmpID[empID] = mar
			}
		}
	}

	type ServicesData struct{ AppointmentDate, PromotionDate, TransferDate string }
	servicesData := map[string]ServicesData{}
	for _, r := range sRows {
		empID := normalizeEmpID(get(r, sm, "Employee ID", "Emp ID"))
		if empID == "" {
			continue
		}
		servicesData[empID] = ServicesData{
			AppointmentDate: get(r, sm, "Date of Appointment"),
			PromotionDate:   get(r, sm, "Last Promotion Order Date"),
			TransferDate:    get(r, sm, "Last Transfer Order Date"),
		}
	}

	// Build EMP
	log.Println("👥 Building employee records...")
	EMP = map[string]Emp{}
	zoneCounts := map[string]int{}
	desigCounts := map[string]int{}

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
		rawRel := get(r, bm, "Religion")
		rel := canonicalReligion(rawRel)

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
			SelectionCategory: catByEmpID[id],
			MaritalStatus:     maritalByEmpID[id],
			Age:               ageFromDOB(dob),
			Mobile:            stripDot0(get(r, bm, "Mobile No.", "Mobile")),
			Email:             emailByEmpID[id],
			DOJ:               dojByEmpID[id],
			FatherName:        get(r, bm, "Father's Name"),
			MotherName:        get(r, bm, "Mother's Name"),
			SpouseName:        get(r, bm, "Spouse Name"),
			Correspondence:    get(r, bm, "Correspondance_Address", "Correspondence Address"),
			Permanent:         get(r, bm, "Permanent_Address", "Permanent Address"),
			HomeTown:          get(r, bm, "Home Town"),
			Religion:          rel,
			AppointmentDate:   servicesData[id].AppointmentDate,
			PromotionDate:     servicesData[id].PromotionDate,
			TransferDate:      servicesData[id].TransferDate,
		}
		EMP[id] = e

		zoneCounts[e.Zone]++
		d := e.Designation
		if d == "" {
			d = "UNKNOWN"
		}
		desigCounts[d]++
	}

	// Build SCH from DBT
	log.Println("🏫 Building school records...")
	SCH = map[string]School{}
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
			ID:                  sid,
			Name:                sname,
			Zone:                zoneName,
			SIName:              get(r, dm, "School Inspector's Name", "SI Name"),
			TotalEnrolment:      atoiSafe(get(r, dm, "Total Enrolment (Last)", "Total Enrolment")),
			MaxEnrolment:        atoiSafe(get(r, dm, "Max Enrolment")),
			MaxEnrolmentDate:    get(r, dm, "Max Enrolment Date"),
			MaxPresent:          atoiSafe(get(r, dm, "Max Present")),
			MaxPresentDate:      get(r, dm, "Max Present Date"),
			WithAccount:         atoiSafe(get(r, dm, "With Account")),
			WithoutAccount:      atoiSafe(get(r, dm, "Without Account")),
			WithAadhaar:         atoiSafe(get(r, dm, "With Aadhaar")),
			WithoutAadhaar:      atoiSafe(get(r, dm, "Without Aadhaar")),
			AadhaarLinkedAcc:    atoiSafe(get(r, dm, "Aadhaar Linked Account")),
			NewAdmissionMonth:   atoiSafe(get(r, dm, "New Admission (This month)")),
			NewAdmissionSession: atoiSafe(get(r, dm, "New Admission (This session)")),
			DBTStudent:          atoiSafe(get(r, dm, "DBT Received (Student)")),
			DBTParent:           atoiSafe(get(r, dm, "DBT Received (Parent)")),
			DBTTotal:            atoiSafe(get(r, dm, "Received By (Student + Parent)")),
		}
		schoolByID[sid] = true
		zoneSet[zoneName] = true
	}
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

	// Totals + charts data
	log.Println("📊 Building chart data...")
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
	ZLBL, ZVAL, DLBL, DVAL = []string{}, []string{}, []string{}, []string{}
	for _, p := range zTop {
		ZLBL = append(ZLBL, p.K)
		ZVAL = append(ZVAL, strconv.Itoa(p.V))
	}
	for _, p := range dTop {
		DLBL = append(DLBL, p.K)
		DVAL = append(DVAL, strconv.Itoa(p.V))
	}

	// Demographics (dynamic categories + religion)
	log.Println("📊 Building demographics...")
	zmap := map[string]map[string]*DemoStats{} // zone -> designation -> stats
	totalMale, totalFemale := 0, 0
	CATEGORY_WISE = map[string]int{}
	for _, e := range EMP {
		if e.Zone == "" || e.Designation == "" {
			continue
		}
		zone := strings.ToUpper(strings.TrimSpace(e.Zone))
		desig := strings.TrimSpace(e.Designation)
		cat := e.SelectionCategory // Already normalized in EMP
		if cat == "" {
			cat = "UNKNOWN"
		}
		rel := e.Religion // Already normalized in EMP
		if rel == "" {
			rel = "UNKNOWN"
		}
		g := strings.ToUpper(strings.TrimSpace(e.Gender))
		isMale, isFemale := g == "MALE" || g == "M", g == "FEMALE" || g == "F"

		if zmap[zone] == nil {
			zmap[zone] = map[string]*DemoStats{}
		}
		if zmap[zone][desig] == nil {
			zmap[zone][desig] = &DemoStats{CatStats: make(map[string]*Stat), ReligionStats: make(map[string]*Stat)}
		}
		stats := zmap[zone][desig]
		if _, ok := stats.CatStats[cat]; !ok {
			stats.CatStats[cat] = &Stat{}
		}
		if isMale {
			stats.CatStats[cat].Male++
			stats.TotalMale++
			totalMale++
		}
		if isFemale {
			stats.CatStats[cat].Female++
			stats.TotalFemale++
			totalFemale++
		}
		stats.Total++
		CATEGORY_WISE[cat]++

		if _, ok := stats.ReligionStats[rel]; !ok {
			stats.ReligionStats[rel] = &Stat{}
		}
		if isMale {
			stats.ReligionStats[rel].Male++
		}
		if isFemale {
			stats.ReligionStats[rel].Female++
		}
		log.Printf("Employee %s: Zone=%s, Desig=%s, Cat=%s, Rel=%s, Male=%v, Female=%v", e.ID, zone, desig, cat, rel, isMale, isFemale)
	}
	GENDER_WISE = map[string]int{"Male": totalMale, "Female": totalFemale}

	// Log demographics for debugging
	for zone, dm := range zmap {
		log.Printf("Zone: %s", zone)
		for dname, st := range dm {
			log.Printf("  Designation: %s, CatStats: %v, ReligionStats: %v, TotalMale: %d, TotalFemale: %d, Total: %d", dname, st.CatStats, st.ReligionStats, st.TotalMale, st.TotalFemale, st.Total)
		}
	}

	DEMO_ZONES = []ZoneDemo{}
	for zone, dm := range zmap {
		var designations []DesignationDemo
		for dname, st := range dm {
			designations = append(designations, DesignationDemo{Designation: dname, Stats: *st})
		}
		sort.Slice(designations, func(i, j int) bool { return designations[i].Designation < designations[j].Designation })
		DEMO_ZONES = append(DEMO_ZONES, ZoneDemo{Zone: zone, Designations: designations})
	}
	sort.Slice(DEMO_ZONES, func(i, j int) bool { return DEMO_ZONES[i].Zone < DEMO_ZONES[j].Zone })
}

// ---- HTTP ----
func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if err := os.MkdirAll(filepath.Dir(*cachePath), 0755); err != nil {
		log.Printf("mkdir cache: %v", err)
	}

	buildAll()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/emp", handleAPIEmployee)
	mux.HandleFunc("/api/school", handleAPISchool)
	mux.HandleFunc("/toggle-live", handleToggleLive)
	mux.HandleFunc("/send-birthdays", handleSendBirthdays)
	mux.HandleFunc("/send-anniversaries", handleSendAnniversaries)
	mux.HandleFunc("/send-whatsapp-invite", handleSendWhatsAppInvite)
	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		buildAll()
		http.Redirect(w, r, "/", http.StatusFound)
	})
	log.Printf("✅ Server running on http://localhost%v", *listen)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := strings.ReplaceAll(indexHTML, "{{TITLE}}", html.EscapeString(*title))

	jsonStr := func(v any) string { b, _ := json.Marshal(v); return string(b) }
	page = strings.ReplaceAll(page, "{{EMP_JSON}}", jsonStr(EMP))
	page = strings.ReplaceAll(page, "{{SCH_JSON}}", jsonStr(SCH))
	page = strings.ReplaceAll(page, "{{DEMO_ZONES_JSON}}", jsonStr(DEMO_ZONES))
	page = strings.ReplaceAll(page, "{{CATEGORY_WISE_JSON}}", jsonStr(CATEGORY_WISE))
	page = strings.ReplaceAll(page, "{{GENDER_WISE_JSON}}", jsonStr(GENDER_WISE))
	page = strings.ReplaceAll(page, "{{ZLBL}}", jsonStr(ZLBL))
	page = strings.ReplaceAll(page, "{{ZVAL}}", jsonStr(ZVAL))
	page = strings.ReplaceAll(page, "{{DLBL}}", jsonStr(DLBL))
	page = strings.ReplaceAll(page, "{{DVAL}}", jsonStr(DVAL))

	page = strings.ReplaceAll(page, "{{TOTAL_EMP}}", strconv.Itoa(len(EMP)))
	page = strings.ReplaceAll(page, "{{TOTAL_SCHOOLS}}", strconv.Itoa(len(SCH)))
	zoneSet := map[string]bool{}
	desigSet := map[string]bool{}
	for _, e := range EMP {
		if e.Zone != "" {
			zoneSet[e.Zone] = true
		}
		if e.Designation != "" {
			desigSet[e.Designation] = true
		}
	}
	page = strings.ReplaceAll(page, "{{TOTAL_ZONES}}", strconv.Itoa(len(zoneSet)))
	page = strings.ReplaceAll(page, "{{TOTAL_DESIGS}}", strconv.Itoa(len(desigSet)))

	_, _ = w.Write([]byte(page))
}

func handleAPIEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := normalizeEmpID(r.URL.Query().Get("id"))
	if id == "" {
		http.Error(w, `{"error":"missing id"}`, 400)
		return
	}
	e, ok := EMP[id]
	if !ok {
		http.Error(w, `{"error":"not found"}`, 404)
		return
	}
	_ = json.NewEncoder(w).Encode(e)
}

func handleAPISchool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := digitsOnly(r.URL.Query().Get("id"))
	if id == "" {
		http.Error(w, `{"error":"missing id"}`, 400)
		return
	}
	s, ok := SCH[id]
	if !ok {
		http.Error(w, `{"error":"not found"}`, 404)
		return
	}

	// attach roster + summary
	type Resp struct {
		School School      `json:"school"`
		Roster []Emp       `json:"roster"`
		Staff  SchoolStaff `json:"staff"`
	}
	roster := []Emp{}
	actualT := 0
	hasP := false
	hasSE := false
	totalSt := 0
	for _, e := range EMP {
		if e.SchoolID == id {
			roster = append(roster, e)
			totalSt++
			d := strings.ToLower(e.Designation)
			if strings.Contains(d, "teacher") {
				actualT++
			}
			if strings.Contains(d, "principal") {
				hasP = true
			}
			if strings.Contains(d, "special educator") {
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
	resp := Resp{
		School: s,
		Roster: roster,
		Staff: SchoolStaff{
			ID: s.ID, Name: s.Name, Zone: s.Zone,
			NeededTeachers: neededT, ActualTeachers: actualT, SurplusVacancy: sv,
			HasPrincipal: hasP, HasSpecialEdu: hasSE, TotalStaff: totalSt, Ratio: ratio,
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleToggleLive(w http.ResponseWriter, r *http.Request) {
	on := r.URL.Query().Get("on")
	if on == "1" {
		LiveMode = true
	} else {
		LiveMode = false
	}
	fmt.Fprintf(w, "Mode: %v", LiveMode)
}

func handleSendBirthdays(w http.ResponseWriter, r *http.Request) {
	today := time.Now()
	count := 0
	validDOB := 0
	for _, e := range EMP {
		if e.Email == "" || e.DOB == "" {
			continue
		}
		d, err := parseDMYFlexible(e.DOB)
		if err != nil {
			continue
		}
		validDOB++
		if d.Day() == today.Day() && d.Month() == today.Month() {
			sal := "Dear Sir/Madam"
			if strings.HasPrefix(strings.ToUpper(e.Gender), "M") {
				sal = "Dear Sir " + e.Name
			}
			if strings.HasPrefix(strings.ToUpper(e.Gender), "F") {
				sal = "Dear Madam " + e.Name
			}
			sub := "🎉 Warm Birthday Wishes from HQ Team IT Education"
			body := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);">
    <div style="background: white; padding: 30px; border-radius: 15px; box-shadow: 0 10px 30px rgba(0,0,0,0.2);">
        <h1 style="color: #667eea; text-align: center; margin-bottom: 20px;">🎉 Happy Birthday! 🎂</h1>
        <p style="font-size: 16px; line-height: 1.6; color: #333;">%s,</p>
        <p style="font-size: 16px; line-height: 1.6; color: #333;">
            Wishing you a very <strong>Happy Birthday</strong> filled with joy, laughter, and wonderful moments!
        </p>
        <p style="font-size: 16px; line-height: 1.6; color: #333;">
            May this special day bring you happiness and may the year ahead be filled with success and good health!
        </p>
        
        <div style="background: linear-gradient(135deg, #25D366 0%%, #128C7E 100%%); padding: 20px; border-radius: 10px; margin: 25px 0; text-align: center;">
            <h3 style="color: white; margin: 0 0 15px 0;">📢 Join Our WhatsApp Channel!</h3>
            <p style="color: white; margin-bottom: 15px; font-size: 14px;">Stay connected with HQ Team IT Education</p>
            <a href="https://whatsapp.com/channel/0029Vb6hLZd1CYoIFek51P0V" 
               style="display: inline-block; background: white; color: #25D366; padding: 12px 30px; 
                      text-decoration: none; border-radius: 25px; font-weight: bold; font-size: 16px;">
                Join Channel Now →
            </a>
        </div>
        
        <hr style="border: none; border-top: 2px solid #eee; margin: 25px 0;">
        <p style="font-size: 14px; color: #666;">
            With warm regards,<br>
            <strong style="color: #667eea;">%s</strong>
        </p>
    </div>
</div>
`, sal, EmailName)
			if err := sendEmail(e.Email, sub, body); err == nil {
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
	for _, e := range EMP {
		if e.Email == "" || e.DOJ == "" {
			continue
		}
		d, err := parseDMYFlexible(e.DOJ)
		if err != nil {
			continue
		}
		validDOJ++
		if d.Day() == today.Day() && d.Month() == today.Month() {
			yrs := today.Year() - d.Year()
			sal := "Dear Sir/Madam"
			if strings.HasPrefix(strings.ToUpper(e.Gender), "M") {
				sal = "Dear Sir " + e.Name
			}
			if strings.HasPrefix(strings.ToUpper(e.Gender), "F") {
				sal = "Dear Madam " + e.Name
			}
			sub := "🏅 Congratulations on your Work Anniversary!"
			body := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%);">
    <div style="background: white; padding: 30px; border-radius: 15px; box-shadow: 0 10px 30px rgba(0,0,0,0.2);">
        <h1 style="color: #f5576c; text-align: center; margin-bottom: 20px;">🏅 Work Anniversary Celebration! 🎊</h1>
        <p style="font-size: 16px; line-height: 1.6; color: #333;">%s,</p>
        <p style="font-size: 16px; line-height: 1.6; color: #333;">
            Congratulations on completing <strong style="color: #f5576c; font-size: 20px;">%d years</strong> of dedicated service with us!
        </p>
        <p style="font-size: 16px; line-height: 1.6; color: #333;">
            Your commitment, hard work, and contributions have been invaluable to our organization. Thank you for your continued excellence!
        </p>
        
        <div style="background: linear-gradient(135deg, #25D366 0%%, #128C7E 100%%); padding: 20px; border-radius: 10px; margin: 25px 0; text-align: center;">
            <h3 style="color: white; margin: 0 0 15px 0;">📢 Join Our WhatsApp Channel!</h3>
            <p style="color: white; margin-bottom: 15px; font-size: 14px;">Stay connected with HQ Team IT Education</p>
            <a href="https://whatsapp.com/channel/0029Vb6hLZd1CYoIFek51P0V" 
               style="display: inline-block; background: white; color: #25D366; padding: 12px 30px; 
                      text-decoration: none; border-radius: 25px; font-weight: bold; font-size: 16px;">
                Join Channel Now →
            </a>
        </div>
        
        <hr style="border: none; border-top: 2px solid #eee; margin: 25px 0;">
        <p style="font-size: 14px; color: #666;">
            With appreciation and best wishes,<br>
            <strong style="color: #f5576c;">%s</strong>
        </p>
    </div>
</div>
`, sal, yrs, EmailName)
			if err := sendEmail(e.Email, sub, body); err == nil {
				count++
			}
		}
	}
	fmt.Fprintf(w, "🏅 Anniversary emails sent to %d employees. (valid DOJ parsed: %d)", count, validDOJ)
}

func handleSendWhatsAppInvite(w http.ResponseWriter, r *http.Request) {
	count := 0
	channelLink := "https://whatsapp.com/channel/0029Vb6hLZd1CYoIFek51P0V"

	for _, e := range EMP {
		if e.Email == "" {
			continue
		}

		sal := "Dear Sir/Madam"
		if strings.HasPrefix(strings.ToUpper(e.Gender), "M") {
			sal = "Dear Sir " + e.Name
		}
		if strings.HasPrefix(strings.ToUpper(e.Gender), "F") {
			sal = "Dear Madam " + e.Name
		}

		sub := "📢 Join HQ Team IT Education WhatsApp Channel"
		body := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background: linear-gradient(135deg, #25D366 0%%, #128C7E 100%%);">
    <div style="background: white; padding: 30px; border-radius: 15px; box-shadow: 0 10px 30px rgba(0,0,0,0.2);">
        <div style="text-align: center; margin-bottom: 20px;">
            <h1 style="color: #25D366; margin: 0;">HQ TEAM IT EDUCATION</h1>
            <p style="color: #666; margin-top: 10px;">WhatsApp Channel</p>
        </div>
        
        <p style="font-size: 16px; line-height: 1.6; color: #333;">%s,</p>
        
        <p style="font-size: 16px; line-height: 1.6; color: #333;">
            We're excited to invite you to join our <strong>official WhatsApp channel</strong>!
        </p>
        
        <div style="background: #f0f9ff; padding: 20px; border-radius: 10px; margin: 20px 0; border-left: 4px solid #25D366;">
            <p style="margin: 0; color: #333; font-size: 15px;">
                📌 Get instant updates<br>
                📌 Important announcements<br>
                📌 News and events<br>
                📌 Direct communication
            </p>
        </div>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" 
               style="display: inline-block; background: linear-gradient(135deg, #25D366 0%%, #128C7E 100%%); 
                      color: white; padding: 15px 40px; text-decoration: none; border-radius: 30px; 
                      font-weight: bold; font-size: 18px; box-shadow: 0 5px 15px rgba(37, 211, 102, 0.3);">
                📱 Join Channel Now
            </a>
        </div>
        
        <p style="font-size: 14px; color: #666; text-align: center; margin-top: 25px;">
            Click the button above or scan the QR code using your WhatsApp camera
        </p>
        
        <hr style="border: none; border-top: 2px solid #eee; margin: 25px 0;">
        <p style="font-size: 14px; color: #666;">
            Best regards,<br>
            <strong style="color: #25D366;">%s</strong>
        </p>
    </div>
</div>
`, sal, channelLink, EmailName)

		if err := sendEmail(e.Email, sub, body); err == nil {
			count++
		}
	}

	fmt.Fprintf(w, "📢 WhatsApp channel invites sent to %d employees.", count)
}

// ---- Embedded HTML ----
var indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>{{TITLE}}</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;600;700&family=Comic+Neue:wght@400;700&display=swap" rel="stylesheet">
  <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
  <style>
    :root{ --bg:#0d1b2a; --card:#1b263b; --muted:#cbd5e1; --text:#f1f5f9; --accent:#00c4b4;
           --ok:#22c55e; --warn:#facc15; --bad:#ef4444; --chip:#334155; --hover:#1e293b;}
    *{box-sizing:border-box}
    html,body{
      margin:0; padding:0; color:var(--text);
      background: radial-gradient(circle at top right,#0d1b2a,#000814);
      font: 15px/1.6 'Poppins','Comic Neue',system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial;
      overflow-x:hidden;
    }
    .container{max-width:1260px;margin:20px auto;padding:0 14px}
    .row{display:grid;grid-template-columns:1fr 1fr;gap:14px}
    @media (max-width:960px){.row{grid-template-columns:1fr}}
    .card{background: rgba(27,38,59,.88);border-radius:16px;padding:14px 14px 16px;box-shadow:0 10px 25px rgba(0,255,255,.18)}
    h1,h2,h3{margin:6px 0 10px;font-weight:700;color:var(--accent)}
    button,.btn{padding:8px 12px;border-radius:10px;border:0;background:#26324d;color:#fff;cursor:pointer}
    .btn.primary{background:linear-gradient(135deg,#00c4b4,#3b82f6)}
    .input,select{width:100%;padding:10px 12px;border-radius:10px;border:1px solid #2b3550;background:#10182a;color:var(--text)}
    .small{font-size:12px;color:var(--muted)}
    .badge{background:#334155;color:#cbd5e1;border-radius:8px;padding:1px 6px;margin-left:6px;font-size:12px}
    .table-wrap{max-height:52vh;overflow:auto;border:1px solid #27304a;border-radius:12px;margin-top:8px}
    .table-scroll{overflow:auto}
    .data-table{width:100%;border-collapse:collapse;font-size:.92rem;min-width:1400px}
    .data-table th,.data-table td{border:1px solid #27304a;padding:8px 10px;text-align:center;white-space:normal;word-break:break-word;font-size:14px;min-width:95px;max-width:120px;letter-spacing:0.2px;color:#e6edf6}
    .data-table th{background:#101828;position:sticky;top:0;z-index:1;color:#cbd5e1;font-weight:600}
    .grid-2{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px}
    .flex{display:flex;gap:8px;flex-wrap:wrap;align-items:center}
    .header-nums{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin:10px 0}
    .headbox{background:#11182a;border:1px solid #27304a;padding:10px;border-radius:12px;text-align:center}
    .charts{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px}
    .chart-card{background:#11182a;border:1px solid #27304a;border-radius:12px;padding:8px}
    .charts .chart{height:300px}
    .charts canvas{width:100% !important; height:300px !important}
    .sticky{position:sticky;top:8px}
    a{color:#7dd3fc;text-decoration:none}
    .toolbar{display:flex;gap:8px;flex-wrap:wrap;margin:6px 0}
    .data-table td:last-child,
    .data-table th:last-child,
    .data-table td:nth-last-child(2),
    .data-table th:nth-last-child(2),
    .data-table td:nth-last-child(3),
    .data-table th:nth-last-child(3){
      background:#1e293b;
      font-weight:700;
      color:#facc15;
    }
    .table-wrap::-webkit-scrollbar{height:10px}
    .table-wrap::-webkit-scrollbar-thumb{background:#334155;border-radius:8px}
  </style>
  <script>
    window.EMP = {{EMP_JSON}};
    window.SCH = {{SCH_JSON}};
    window.DEMO_ZONES = {{DEMO_ZONES_JSON}};
    window.CATEGORY_WISE = {{CATEGORY_WISE_JSON}};
    window.GENDER_WISE = {{GENDER_WISE_JSON}};
    window.ZLBL = {{ZLBL}}; window.ZVAL = {{ZVAL}}; window.DLBL = {{DLBL}}; window.DVAL = {{DVAL}};
  </script>
</head>
<body>
  <div class="container">
    <h1>🧭 {{TITLE}}</h1>

    <div class="header-nums">
      <div class="headbox"><div class="small">TOTAL EMPLOYEES</div><div style="font-size:26px" id="num-emps">{{TOTAL_EMP}}</div></div>
      <div class="headbox"><div class="small">SCHOOLS</div><div style="font-size:26px" id="num-schools">{{TOTAL_SCHOOLS}}</div></div>
      <div class="headbox"><div class="small">ZONES</div><div style="font-size:26px" id="num-zones">{{TOTAL_ZONES}}</div></div>
      <div class="headbox"><div class="small">DESIGNATIONS</div><div style="font-size:26px" id="num-desigs">{{TOTAL_DESIGS}}</div></div>
    </div>

    <!-- Row 0: Charts Top (2x2) -->
    <div class="charts" id="chartsTop">
      <div class="chart-card"><h3>Zone-wise Distribution</h3><canvas id="zoneChart" class="chart"></canvas></div>
      <div class="chart-card"><h3>Top Designations</h3><canvas id="desigChart" class="chart"></canvas></div>
      <div class="chart-card"><h3>Overall Category (Pie)</h3><canvas id="catChart" class="chart"></canvas></div>
      <div class="chart-card"><h3>Gender Split</h3><canvas id="genderChart" class="chart"></canvas></div>
    </div>

    <!-- Row 1: Lookups -->
    <div class="row" style="margin-top:12px">
      <div class="card">
        <h2>👤 Employee Lookup</h2>
        <div class="flex">
          <input id="empid" class="input" placeholder="Employee ID or Name (min 3)" />
          <button class="btn primary" onclick="findEmp()">Search</button>
        </div>
        <div id="emp-out" style="margin-top:8px"></div>
        <h3>Search by Name</h3>
        <input id="empname" class="input" placeholder="e.g., jyoti, amit" oninput="searchEmpByName()">
        <div class="list" id="empname-results"></div>
      </div>

      <div class="card">
        <h2>🏫 School Lookup</h2>
        <div class="flex">
          <input id="schid" class="input" placeholder="School ID (digits)" />
          <button class="btn primary" onclick="findSchool()">Search</button>
        </div>
        <div id="sch-out" style="margin-top:8px"></div>
        <h3>Search by School Name</h3>
        <input id="schname" class="input" placeholder="e.g., rohini" oninput="searchSchoolByName()">
        <div class="list" id="schname-results"></div>
      </div>
    </div>

    <!-- Row 2: Advanced filter (left) -->
    <div class="row">
      <div class="card sticky">
        <h2>🔎 Advanced Employee Filter</h2>
        <div class="grid-2">
          <div><label class="small">Zone</label><select id="f-zone"></select></div>
          <div><label class="small">Designation</label><select id="f-desig"></select></div>
          <div><label class="small">Gender</label><select id="f-gender"><option>ALL</option><option>MALE</option><option>FEMALE</option></select></div>
          <div><label class="small">Category</label><select id="f-cat"></select></div>
          <div><label class="small">Religion</label><select id="f-rel"></select></div>
          <div><label class="small">Marital Status</label><select id="f-marital"><option>ALL</option><option>Married</option><option>Widow/Widower</option><option>Divorced</option><option>Unmarried</option><option>Separated</option><option>Live-in</option></select></div>
          <div><label class="small">Age ≥</label><select id="f-age"><option>0</option><option>25</option><option>30</option><option>35</option><option>40</option><option>45</option><option>50</option><option>55</option></select></div>
        </div>
        <div class="flex" style="margin-top:8px">
          <button class="btn primary" onclick="runEmpFilter()">Apply</button>
          <button class="btn" onclick="resetEmpFilter()">Reset</button>
          <button class="btn" onclick="exportTableCSV('empFilterTable','employees_filtered.csv')">Export CSV</button>
        </div>
        <div class="table-wrap table-scroll" style="margin-top:8px">
          <table class="data-table" id="empFilterTable">
            <thead>
              <tr><th>Employee</th><th>Designation</th><th>Zone</th><th>Gender</th><th>Category</th><th>Religion</th><th>Marital</th><th>Age</th><th>School</th></tr>
            </thead>
            <tbody></tbody>
          </table>
        </div>
      </div>

      <div class="card">
        <h2>Info</h2>
        <div class="small">Use Advanced Filter to slice the employee list instantly. Charts above reflect overall data (not filtered).</div>
      </div>
    </div>

    <!-- Row 3: Demographic tables with their own filters -->
    <div class="card">
      <div class="flex" style="justify-content:space-between; align-items:flex-end">
        <div>
          <h2>📋 Zone-wise Demographic Summary (Dynamic Categories)</h2>
          <div class="toolbar">
            <select id="demo-zone"></select>
            <select id="demo-desig"></select>
            <select id="demo-cat"></select>
            <button class="btn" onclick="applyDemoFilter()">Apply</button>
            <button class="btn" onclick="resetDemoFilter()">Reset</button>
          </div>
        </div>
        <div class="flex">
          <button class="btn" onclick="exportTableCSV('demoTable','demographics_zone_summary.csv')">Export CSV</button>
        </div>
      </div>
      <div class="table-wrap table-scroll">
        <table class="data-table" id="demoTable">
          <thead><tr><th>Zone</th><th>Designation</th></tr></thead>
          <tbody></tbody>
        </table>
      </div>
    </div>

    <div class="card">
      <div class="flex" style="justify-content:space-between; align-items:flex-end">
        <div>
          <h2>🕌 Religion-wise Staff Distribution</h2>
          <div class="toolbar">
            <select id="rel-zone"></select>
            <select id="rel-desig"></select>
            <select id="rel-rel"></select>
            <button class="btn" onclick="applyReligionFilter()">Apply</button>
            <button class="btn" onclick="resetReligionFilter()">Reset</button>
          </div>
        </div>
        <button class="btn" onclick="exportTableCSV('religionTable','religion_summary.csv')">Export CSV</button>
      </div>
      <div class="table-wrap table-scroll">
        <table class="data-table" id="religionTable">
          <thead><tr><th>Zone</th><th>Designation</th></tr></thead>
          <tbody></tbody>
        </table>
      </div>
    </div>

    <div class="flex" style="margin-top:10px">
      <a class="btn" href="/refresh">🔄 Refresh</a>
      <a class="btn" href="/toggle-live?on=1">✉️ Enable Live Email</a>
      <a class="btn" href="/toggle-live?on=0">✉️ Disable Live Email</a>
      <a class="btn" href="/send-birthdays">🎂 Send Birthdays</a>
      <a class="btn" href="/send-anniversaries">🏅 Send Anniversaries</a>
	  <a class="btn" href="/send-whatsapp-invite">📢 Send WhatsApp Invite to All</a>
    </div>

    <button onclick="window.scrollTo({top:0,behavior:'smooth'})" style="position:fixed;right:14px;bottom:16px" class="btn">⬆ Top</button>

    <div class="small" style="margin-top:6px">Tip: Table CSV export respects current filters.</div>
  </div>
<div class="card" style="margin-top:12px">
  <h2>📢 WhatsApp Channel</h2>
  <div style="text-align:center; padding:20px">
    <div style="background: linear-gradient(135deg, #25D366 0%, #128C7E 100%); 
                padding: 30px; border-radius: 15px; color: white;">
      <h3 style="margin-top:0; color:white">HQ TEAM IT EDUCATION</h3>
      <p style="margin: 15px 0">Join our WhatsApp channel for updates and announcements</p>
      
      <div style="background: white; padding: 20px; border-radius: 10px; margin: 20px auto; max-width: 300px;">
        <img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAASwAAAEsCAYAAAB5fY51AAAMYElEQVR4nO3dQW4bRxaA4T..." 
             alt="QR Code" style="width: 250px; height: 250px;">
      </div>
      
      <a href="https://whatsapp.com/channel/0029Vb6hLZd1CYoIFek51P0V" 
         class="btn" 
         style="background: white; color: #25D366; font-weight: bold; padding: 15px 40px; 
                font-size: 18px; text-decoration: none; display: inline-block; margin-top: 15px;">
        📱 Join Channel Now
      </a>
      
      <p style="margin-top: 20px; font-size: 14px; opacity: 0.9;">
        Or scan the QR code with your WhatsApp camera
      </p>
    </div>
  </div>
</div>
  <script>
    var EMP = (typeof window.EMP==='string')?JSON.parse(window.EMP):(window.EMP||{});
    var SCH = (typeof window.SCH==='string')?JSON.parse(window.SCH):(window.SCH||{});
    var DEMO_ZONES = (typeof window.DEMO_ZONES==='string')?JSON.parse(window.DEMO_ZONES):(window.DEMO_ZONES||[]);
    var CATEGORY_WISE = (typeof window.CATEGORY_WISE==='string')?JSON.parse(window.CATEGORY_WISE):(window.CATEGORY_WISE||{});
    var GENDER_WISE = (typeof window.GENDER_WISE==='string')?JSON.parse(window.GENDER_WISE):(window.GENDER_WISE||{});
    window.ZLBL = {{ZLBL}}; window.ZVAL = {{ZVAL}}; window.DLBL = {{DLBL}}; window.DVAL = {{DVAL}};

    function fmt(n){ return (n==null?0:n).toLocaleString(); }
    function digitsOnlyKey(s){ return (s||'').replace(/\\D+/g,''); }

    // Charts
    (function(){
      try{
        var zctx=document.getElementById('zoneChart').getContext('2d');
        new Chart(zctx,{
          type:'bar',
          data:{
            labels:(window.ZLBL||[]),
            datasets:[{
              label:'Employees',
              data:(window.ZVAL||[]),
              backgroundColor:'rgba(0,196,180,0.7)',
              borderColor:'rgba(0,196,180,1)',
              borderWidth:1
            }]
          },
          options:{scales:{y:{beginAtZero:true}}}
        });
        var dctx=document.getElementById('desigChart').getContext('2d');
        new Chart(dctx,{
          type:'bar',
          data:{
            labels:(window.DLBL||[]),
            datasets:[{
              label:'Employees',
              data:(window.DVAL||[]),
              backgroundColor:'rgba(59,130,246,0.7)',
              borderColor:'rgba(59,130,246,1)',
              borderWidth:1
            }]
          },
          options:{scales:{y:{beginAtZero:true}}}
        });
        var cctx=document.getElementById('catChart').getContext('2d');
        var catLabels=Object.keys(CATEGORY_WISE);
        var catVals=catLabels.map(function(k){return CATEGORY_WISE[k];});
        new Chart(cctx,{
          type:'pie',
          data:{
            labels:catLabels,
            datasets:[{
              data:catVals,
              backgroundColor:['#00c4b4','#3b82f6','#22c55e','#facc15','#ef4444','#a855f7']
            }]
          }
        });
        var gctx=document.getElementById('genderChart').getContext('2d');
        var gLabels=Object.keys(GENDER_WISE);
        var gVals=gLabels.map(function(k){return GENDER_WISE[k];});
        new Chart(gctx,{
          type:'doughnut',
          data:{
            labels:gLabels,
            datasets:[{
              data:gVals,
              backgroundColor:['#3b82f6','#f472b6']
            }]
          }
        });
      }catch(e){ console.warn('chart init failed', e); }
    })();

    // Advanced Filter options
    function initEmpFilterOptions(){
      var zones=new Set(["ALL"]); var desigs=new Set(["ALL"]); var cats=new Set(["ALL"]); var rels=new Set(["ALL"]);
      Object.values(EMP).forEach(function(e){
        if(e.zone) zones.add(e.zone);
        if(e.designation) desigs.add(e.designation);
        if(e.selection_category) cats.add(e.selection_category);
        if(e.religion) rels.add(e.religion);
      });
      function fill(id,set){
        var sel=document.getElementById(id); if(!sel) return;
        sel.innerHTML='';
        Array.from(set).sort().forEach(function(v){
          sel.insertAdjacentHTML('beforeend','<option value="'+v+'">'+v+'</option>');
        });
      }
      fill('f-zone', zones); fill('f-desig', desigs); fill('f-cat', cats); fill('f-rel', rels);
      document.getElementById('f-zone').value='ALL';
      document.getElementById('f-desig').value='ALL';
      document.getElementById('f-cat').value='ALL';
      document.getElementById('f-rel').value='ALL';
    }
    document.addEventListener('DOMContentLoaded', initEmpFilterOptions);

    function runEmpFilter(){
      var z=(document.getElementById('f-zone').value||'ALL').toUpperCase();
      var d=(document.getElementById('f-desig').value||'ALL').toUpperCase();
      var g=(document.getElementById('f-gender').value||'ALL').toUpperCase();
      var c=(document.getElementById('f-cat').value||'ALL').toUpperCase();
      var r=(document.getElementById('f-rel').value||'ALL').toUpperCase();
      var m=(document.getElementById('f-marital').value||'ALL').toUpperCase();
      var a=parseInt(document.getElementById('f-age').value||'0',10);
      function mcanon(x){
        x = String(x||'').toUpperCase();
        if(x==='WIDOW' || x==='WIDOWER') return 'WIDOW/WIDOWER';
        if(x==='LIVE IN' || x==='LIVE-IN') return 'LIVE-IN';
        return x;
      }
      m = mcanon(m);

      var tbody=document.querySelector('#empFilterTable tbody'); tbody.innerHTML='';
      var list=[];
      Object.values(EMP).forEach(function(e){
        if(z!=='ALL' && String(e.zone||'').toUpperCase()!==z) return;
        if(d!=='ALL' && String(e.designation||'').toUpperCase()!==d) return;
        var eg=String(e.gender||'').toUpperCase();
        if(g!=='ALL' && eg!==g) return;
        if(c!=='ALL' && String(e.selection_category||'').toUpperCase()!==c) return;
        if(r!=='ALL' && String(e.religion||'').toUpperCase()!==r) return;
        if(m!=='ALL' && mcanon(String(e.marital_status||'').toUpperCase())!==m) return;
        if(a>0 && (e.age||0)<a) return;
        list.push(e);
      });
      list.sort(function(a,b){ return (a.name||'').localeCompare(b.name||''); });
      var out='';
      list.slice(0,500).forEach(function(e){
        out += '<tr>' +
          '<td>' + (e.name||'') + ' <span class="small">(' + (e.id||'') + ')</span></td>' +
          '<td>' + (e.designation||'') + '</td>' +
          '<td>' + (e.zone||'') + '</td>' +
          '<td>' + (e.gender||'') + '</td>' +
          '<td>' + (e.selection_category||'') + '</td>' +
          '<td>' + (e.religion||'') + '</td>' +
          '<td>' + (e.marital_status||'') + '</td>' +
          '<td>' + (e.age||'') + '</td>' +
          '<td>' + (e.school_name||'') + '</td>' +
        '</tr>';
      });
      tbody.innerHTML = out || '<tr><td colspan="9" class="small">No results</td></tr>';
    }
    function resetEmpFilter(){
      initEmpFilterOptions();
      document.getElementById('f-gender').value='ALL';
      document.getElementById('f-marital').value='ALL';
      document.getElementById('f-age').value='0';
      runEmpFilter();
    }
    document.addEventListener('DOMContentLoaded', runEmpFilter);

    // Employee Lookup
    function robustEmpFind(raw){
      if(!raw) return null;
      if(EMP[raw]) return EMP[raw];
      var only = raw.replace(/\\D+/g,'');
      if(only){
        var hitK = Object.keys(EMP).find(function(k){ return k.replace(/\\D+/g,'') === only; });
        if(hitK) return EMP[hitK];
      }
      if(raw.length >= 3){
        var hit = Object.values(EMP).find(function(x){ return (x.name||'').toLowerCase().includes(raw.toLowerCase()); });
        if(hit) return hit;
      }
      return null;
    }
    function findEmp(){
      var val=(document.getElementById('empid').value||'').trim();
      var out=document.getElementById('emp-out');
      if(!val){ out.innerHTML='<div class="small">Please enter Employee ID or Name.</div>'; return; }
      var e = robustEmpFind(val);
      if(!e){ out.innerHTML='<div class="small">No employee found.</div>'; return; }
      out.innerHTML =
        '<details open>' +
          '<summary>📋 Personal Info — ' + (e.name||'') + ' (' + (e.id||'') + ')</summary>' +
          '<div class="grid-2">' +
            '<div><b>Gender:</b> ' + (e.gender||'') + '</div>' +
            '<div><b>DOB:</b> ' + (e.dob||'') + ' <span class="small">(Age ' + (e.age||'') + ')</span></div>' +
            '<div><b>Category:</b> ' + (e.selection_category||'') + '</div>' +
            '<div><b>Religion:</b> ' + (e.religion||'') + '</div>' +
            '<div><b>Marital:</b> ' + (e.marital_status||'') + '</div>' +
            '<div><b>Mobile:</b> ' + (e.mobile||'') + '</div>' +
            '<div><b>Email:</b> ' + (e.email||'') + '</div>' +
          '</div>' +
        '</details>' +
        '<details>' +
          '<summary>🧾 Service Info</summary>' +
          '<div class="grid-2">' +
            '<div><b>Designation:</b> ' + (e.designation||'') + '</div>' +
            '<div><b>DOJ:</b> ' + (e.doj||'') + '</div>' +
            '<div><b>Appointment Date:</b> ' + (e.appointment_date||'') + '</div>' +
            '<div><b>Promotion Date:</b> ' + (e.promotion_date||'') + '</div>' +
            '<div><b>Transfer Date:</b> ' + (e.transfer_date||'') + '</div>' +
            '<div><b>Zone:</b> ' + (e.zone||'') + '</div>' +
            '<div><b>School:</b> ' + (e.school_name||'') + '</div>' +
          '</div>' +
        '</details>' +
        '<details>' +
          '<summary>🏠 Address & Family</summary>' +
          '<div class="grid-2">' +
            '<div><b>Father:</b> ' + (e.father_name||'') + '</div>' +
            '<div><b>Mother:</b> ' + (e.mother_name||'') + '</div>' +
            '<div><b>Spouse:</b> ' + (e.spouse_name||'') + '</div>' +
            '<div><b>Correspondence:</b> ' + (e.correspondence||'') + '</div>' +
            '<div><b>Permanent:</b> ' + (e.permanent||'') + '</div>' +
            '<div><b>Home Town:</b> ' + (e.home_town||'') + '</div>' +
          '</div>' +
        '</details>';
    }

    // School Lookup
    function findSchool(){
      var raw=(document.getElementById('schid').value||''); var sid=digitsOnlyKey(raw);
      var out=document.getElementById('sch-out');
      if(!sid || !SCH[sid]){ out.innerHTML='<div class="small">No school found.</div>'; return; }
      var s = SCH[sid];
      var roster = Object.values(EMP).filter(function(e){ return e.school_id === sid; });
      var priority = ["principal", "special educator", "teacher (primary)"];
      roster.sort(function(a,b){
        var da=(a.designation||'').toLowerCase(); var db=(b.designation||'').toLowerCase();
        var ia=priority.findIndex(function(p){ return da.includes(p); });
        var ib=priority.findIndex(function(p){ return db.includes(p); });
        if(ia===-1 && ib===-1) return da.localeCompare(db);
        if(ia===-1) return 1; if(ib===-1) return -1; return ia-ib;
      });
      var desigCount = {}; roster.forEach(function(e){ var d=(e.designation||'Unknown').trim(); desigCount[d]=(desigCount[d]||0)+1; });
      var desigRows = Object.keys(desigCount).sort().map(function(d){
        return '<tr><td>'+d+'</td><td>'+desigCount[d]+'</td></tr>';
      }).join('');
      var teacherCount = roster.filter(function(e){ return (e.designation||'').toLowerCase().includes('teacher'); }).length;
      var hasPrincipal = roster.some(function(e){ return (e.designation||'').toLowerCase().includes('principal'); });
      var hasSE = roster.some(function(e){ return (e.designation||'').toLowerCase().includes('special educator'); });
      var needed = Math.ceil((s.max_present||0)/40);
      var surplus = teacherCount - needed;

      out.innerHTML =
        '<details open>' +
          '<summary>🏫 School Info — ' + (s.name||'') + ' (' + (s.id||'') + ')</summary>' +
          '<div class="grid-2">' +
            '<div><b>Zone:</b> ' + (s.zone||'') + '</div>' +
            '<div><b>Inspector:</b> ' + (s.si_name||'') + '</div>' +
            '<div><b>Total Enrolment:</b> ' + fmt(s.total_enrolment||0) + '</div>' +
            '<div><b>Max Enrolment:</b> ' + fmt(s.max_enrolment||0) + ' ' + (s.max_enrolment_date?('('+s.max_enrolment_date+')'):'') + '</div>' +
            '<div><b>Max Present:</b> ' + fmt(s.max_present||0) + ' ' + (s.max_present_date?('('+s.max_present_date+')'):'') + '</div>' +
            '<div><b>With Account:</b> ' + fmt(s.with_account||0) + '</div>' +
            '<div><b>Without Account:</b> ' + fmt(s.without_account||0) + '</div>' +
            '<div><b>With Aadhaar:</b> ' + fmt(s.with_aadhaar||0) + '</div>' +
            '<div><b>Without Aadhaar:</b> ' + fmt(s.without_aadhaar||0) + '</div>' +
            '<div><b>Aadhaar Linked Account:</b> ' + fmt(s.aadhaar_linked_acc||0) + '</div>' +
            '<div><b>New Admission (This Month):</b> ' + fmt(s.new_admission_month||0) + '</div>' +
            '<div><b>New Admission (This Session):</b> ' + fmt(s.new_admission_session||0) + '</div>' +
            '<div><b>DBT Received (Student):</b> ' + fmt(s.dbt_student||0) + '</div>' +
            '<div><b>DBT Received (Parent):</b> ' + fmt(s.dbt_parent||0) + '</div>' +
            '<div><b>Total Received (Student + Parent):</b> ' + fmt(s.dbt_total||0) + '</div>' +
          '</div>' +
        '</details>' +
        '<details>' +
          '<summary>👩‍🏫 Staff & Ratio</summary>' +
          '<div class="grid-2">' +
            '<div><b>Teachers:</b> ' + fmt(teacherCount) + '</div>' +
            '<div><b>Needed (1:40):</b> ' + fmt(needed) + '</div>' +
            '<div><b>Status:</b> ' + (surplus>0?('Surplus '+fmt(surplus)):surplus<0?('Vacancy '+fmt(-surplus)):'Balanced') + '</div>' +
            '<div><b>Principal:</b> ' + (hasPrincipal?'Available':'Vacant') + '</div>' +
            '<div><b>Special Educator:</b> ' + (hasSE?'Available':'Vacant') + '</div>' +
            '<div><b>Total Staff:</b> ' + fmt(roster.length) + '</div>' +
          '</div>' +
          '<div style="margin-top:8px" class="table-wrap">' +
            '<table class="data-table"><thead><tr><th>Designation</th><th>Count</th></tr></thead><tbody>' + desigRows + '</tbody></table>' +
          '</div>' +
        '</details>' +
        '<details>' +
          '<summary>🧾 Staff Roster (' + roster.length + ')</summary>' +
          '<div class="table-wrap">' +
            '<table class="data-table">' +
              '<thead><tr><th>Emp ID</th><th>Name</th><th>Designation</th><th>Gender</th><th>Category</th><th>DOJ</th></tr></thead>' +
              '<tbody>' +
                roster.map(function(e){ return '<tr><td>'+e.id+'</td><td>'+e.name+'</td><td>'+e.designation+'</td><td>'+e.gender+'</td><td>'+(e.selection_category||'')+'</td><td>'+(e.doj||'')+'</td></tr>'; }).join('') +
              '</tbody>' +
            '</table>' +
          '</div>' +
        '</details>';
    }

    // Name search helpers
    function searchSchoolByName(){
      var q=(document.getElementById('schname')||{}).value||''; var wrap=document.getElementById('schname-results'); if(!wrap) return;
      wrap.innerHTML=''; var qq=q.trim().toLowerCase(); if(!qq) return;
      var arr=[]; for(var sid in SCH){ var s = SCH[sid]; if(!s||!s.name) continue; if(s.name.toLowerCase().includes(qq)){ arr.push(s); } }
      if(!arr.length){ wrap.innerHTML='<div class="small">No matching schools</div>'; return; }
      arr.sort(function(a,b){ return a.name.localeCompare(b.name); });
      wrap.innerHTML = arr.slice(0,50).map(function(s){
        return '<div class="item" onclick="document.getElementById(\'schid\').value=\''+s.id+'\'; findSchool(); window.scrollTo({top:0,behavior:\'smooth\'})">' +
          '<div><b>'+s.name+'</b></div><div class="small">'+s.id+' • '+(s.zone||'')+'</div></div>';
      }).join('');
    }
    function searchEmpByName(){
      var q=(document.getElementById('empname')||{}).value||''; var wrap=document.getElementById('empname-results'); if(!wrap) return;
      wrap.innerHTML=''; var qq=q.trim().toLowerCase(); if(!qq) return;
      var arr=[]; for(var id in EMP){ var e = EMP[id]; if(!e||!e.name) continue; if(e.name.toLowerCase().includes(qq)){ arr.push(e); } }
      if(!arr.length){ wrap.innerHTML='<div class="small">No matching employees</div>'; return; }
      arr.sort(function(a,b){ return a.name.localeCompare(b.name); });
      wrap.innerHTML = arr.slice(0,50).map(function(e){
        return '<div class="item" onclick="document.getElementById(\'empid\').value=\''+e.id+'\'; findEmp(); window.scrollTo({top:0,behavior:\'smooth\'})">' +
          '<div><b>'+e.name+'</b> <span class="badge">'+(e.designation||'')+'</span></div>' +
          '<div class="small">'+e.id+' • '+(e.school_name||'')+'</div></div>';
      }).join('');
    }

    // ===== Demographic Table (with filters) =====
    var DEMO_CAT_HEADERS = [];
    function collectDemoHeaders(){
      var cats = new Set();
      DEMO_ZONES.forEach(function(z){
        (z.designations||[]).forEach(function(d){
          var cs = (d.stats && d.stats.cat_stats) ? d.stats.cat_stats : {};
          Object.keys(cs).forEach(function(k){ cats.add(k); });
        });
      });
      DEMO_CAT_HEADERS = Array.from(cats).sort();
    }

    function fillDemoFilters(){
      var zset=new Set(["ALL ZONES"]);
      var dset=new Set(["ALL DESIGNATIONS"]);
      var cset=new Set(["ALL CATEGORIES"]);
      DEMO_ZONES.forEach(function(z){
        zset.add(z.zone);
        (z.designations||[]).forEach(function(d){
          dset.add(d.designation);
          var cs = (d.stats && d.stats.cat_stats) ? d.stats.cat_stats : {};
          Object.keys(cs).forEach(function(k){ cset.add(k); });
        });
      });
      function fill(id,set){
        var sel=document.getElementById(id); if(!sel) return;
        sel.innerHTML='';
        Array.from(set).sort().forEach(function(v){ sel.insertAdjacentHTML('beforeend','<option>'+v+'</option>'); });
      }
      fill('demo-zone', zset); fill('demo-desig', dset); fill('demo-cat', cset);
      document.getElementById('demo-zone').value='ALL ZONES';
      document.getElementById('demo-desig').value='ALL DESIGNATIONS';
      document.getElementById('demo-cat').value='ALL CATEGORIES';
    }

    function buildDemoTable(){
      collectDemoHeaders();
      var head = document.querySelector("#demoTable thead");
      var body = document.querySelector("#demoTable tbody");
      if(!head || !body) return;

      var hdr = '<tr><th>Zone</th><th>Designation</th>';
      DEMO_CAT_HEADERS.forEach(function(c){ hdr += '<th>'+c+' ♂</th><th>'+c+' ♀</th>'; });
      hdr += '<th>Total Male</th><th>Total Female</th><th>Total</th></tr>';
      head.innerHTML = hdr;

      renderDemoBody();
    }

    function applyDemoFilter(){ renderDemoBody(); }
    function resetDemoFilter(){ fillDemoFilters(); renderDemoBody(); }

    function renderDemoBody(){
      var body = document.querySelector("#demoTable tbody"); if(!body) return;
      var fz = (document.getElementById('demo-zone').value||'ALL ZONES').toUpperCase();
      var fd = (document.getElementById('demo-desig').value||'ALL DESIGNATIONS').toUpperCase();
      var fc = (document.getElementById('demo-cat').value||'ALL CATEGORIES').toUpperCase();

      var rows='';
      DEMO_ZONES.forEach(function(z){
        if(fz!=='ALL ZONES' && String(z.zone||'').toUpperCase()!==fz) return;
        (z.designations||[]).forEach(function(d){
          if(fd!=='ALL DESIGNATIONS' && String(d.designation||'').toUpperCase()!==fd) return;
          var s = d.stats || {}; var cs = s.cat_stats || {};
          if(fc!=='ALL CATEGORIES' && !cs[fc]){ return; }
          var cells = '<td>'+z.zone+'</td><td>'+d.designation+'</td>';
          DEMO_CAT_HEADERS.forEach(function(c){
            var v = cs[c] || {male:0, female:0};
            var m = (v.male==null?0:v.male), f=(v.female==null?0:v.female);
            cells += '<td>'+ m +'</td><td>'+ f +'</td>';
          });
          cells += '<td>'+ (s.total_male||0) +'</td><td>'+ (s.total_female||0) +'</td><td>'+ (s.total||0) +'</td>';
          rows += '<tr>'+cells+'</tr>';
        });
      });
      body.innerHTML = rows || '<tr><td colspan="'+(DEMO_CAT_HEADERS.length*2+3)+'" class="small">No data</td></tr>';
    }

    // ===== Religion Table (with filters) =====
    var REL_HEADERS = [];
    function collectReligionHeaders(){
      var rels = new Set();
      DEMO_ZONES.forEach(function(z){
        (z.designations||[]).forEach(function(d){
          var rs = (d.stats && d.stats.religion_stats) ? d.stats.religion_stats : {};
          Object.keys(rs).forEach(function(k){ rels.add(k); });
        });
      });
      REL_HEADERS = Array.from(rels).sort();
    }

    function fillReligionFilters(){
      var zset=new Set(["ALL ZONES"]);
      var dset=new Set(["ALL DESIGNATIONS"]);
      var rset=new Set(["ALL RELIGIONS"]);
      DEMO_ZONES.forEach(function(z){
        zset.add(z.zone);
        (z.designations||[]).forEach(function(d){
          dset.add(d.designation);
          var rs = (d.stats && d.stats.religion_stats) ? d.stats.religion_stats : {};
          Object.keys(rs).forEach(function(k){ rset.add(k); });
        });
      });
      function fill(id,set){
        var sel=document.getElementById(id); if(!sel) return;
        sel.innerHTML='';
        Array.from(set).sort().forEach(function(v){ sel.insertAdjacentHTML('beforeend','<option>'+v+'</option>'); });
      }
      fill('rel-zone', zset); fill('rel-desig', dset); fill('rel-rel', rset);
      document.getElementById('rel-zone').value='ALL ZONES';
      document.getElementById('rel-desig').value='ALL DESIGNATIONS';
      document.getElementById('rel-rel').value='ALL RELIGIONS';
    }

    function buildReligionTable(){
      collectReligionHeaders();
      var head = document.querySelector("#religionTable thead");
      var body = document.querySelector("#religionTable tbody");
      if(!head || !body) return;

      var hdr = '<tr><th>Zone</th><th>Designation</th>';
      REL_HEADERS.forEach(function(r){ hdr += '<th>'+r+' ♂</th><th>'+r+' ♀</th>'; });
      hdr += '<th>Total Male</th><th>Total Female</th><th>Total</th></tr>';
      head.innerHTML = hdr;

      renderReligionBody();
    }

    function applyReligionFilter(){ renderReligionBody(); }
    function resetReligionFilter(){ fillReligionFilters(); renderReligionBody(); }

    function renderReligionBody(){
      var body = document.querySelector("#religionTable tbody"); if(!body) return;
      var fz = (document.getElementById('rel-zone').value||'ALL ZONES').toUpperCase();
      var fd = (document.getElementById('rel-desig').value||'ALL DESIGNATIONS').toUpperCase();
      var fr = (document.getElementById('rel-rel').value||'ALL RELIGIONS').toUpperCase();

      var rows='';
      DEMO_ZONES.forEach(function(z){
        if(fz!=='ALL ZONES' && String(z.zone||'').toUpperCase()!==fz) return;
        (z.designations||[]).forEach(function(d){
          if(fd!=='ALL DESIGNATIONS' && String(d.designation||'').toUpperCase()!==fd) return;
          var s = d.stats || {}; var rs = s.religion_stats || {};
          if(fr!=='ALL RELIGIONS' && !rs[fr]){ return; }
          var cells = '<td>'+z.zone+'</td><td>'+d.designation+'</td>';
          REL_HEADERS.forEach(function(r){
            var v = rs[r] || {male:0, female:0};
            var m = (v.male==null?0:v.male), f=(v.female==null?0:v.female);
            cells += '<td>'+ m +'</td><td>'+ f +'</td>';
          });
          cells += '<td>'+ (s.total_male||0) +'</td><td>'+ (s.total_female||0) +'</td><td>'+ (s.total||0) +'</td>';
          rows += '<tr>'+cells+'</tr>';
        });
      });
      body.innerHTML = rows || '<tr><td colspan="'+(REL_HEADERS.length*2+3)+'" class="small">No data</td></tr>';
    }

    // CSV Export
    function exportTableCSV(tableId, filename){
      var table = document.getElementById(tableId); if(!table) return;
      var rows = Array.from(table.querySelectorAll('tr'));
      var csv = rows.map(function(row){
        return Array.from(row.querySelectorAll('th,td')).map(function(cell){
          return '"' + (cell.textContent||'').replace(/"/g,'""') + '"';
        }).join(',');
      }).join('\n');
      var blob = new Blob([csv], {type: 'text/csv'});
      var url = window.URL.createObjectURL(blob);
      var a = document.createElement('a'); a.href = url; a.download = filename;
      a.click(); window.URL.revokeObjectURL(url);
    }

    // Initialize tables
    document.addEventListener('DOMContentLoaded', function(){
      buildDemoTable();
      buildReligionTable();
      fillDemoFilters();
      fillReligionFilters();
    });
  </script>
</body>
</html>`
