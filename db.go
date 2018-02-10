package ip2proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/juju/errors"
)

// DB holds a parsed database instance
type DB struct {
	data        []byte
	dataSize    uint32
	header      *dbHeader
	positions   *positions
	ipv4Indexes [maxIndexes][2]uint32
}

// Result holds the lookup results
type Result struct {
	IP          string
	Country     *string
	CountryCode *string
	City        *string
	ISP         *string
	Region      *string
	Proxy       ProxyType
}

// Database header
type dbHeader struct {
	Count          uint32
	BaseAddr       uint32
	IndexBaseAddr  uint32
	Type           DbType
	Cols           uint8
	Year           uint16
	Month          uint8
	Day            uint8
	IPv4ColumnSize uint8
}

// fields positions according to db type
type positions struct {
	Country uint8
	Region  uint8
	City    uint8
	ISP     uint8
	Proxy   uint8
}

// Open will opens a db file and parses it
func Open(path string) (*DB, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil || len(data) == 0 {
		if err == nil {
			err = fmt.Errorf("%s is empty or not redable", path)
		}
		return nil, errors.Annotate(err, "cannot open/read db file")
	}
	db := &DB{
		data:     data,
		dataSize: uint32(len(data)),
	}

	if err = db.readHeader(); err != nil {
		return nil, errors.Annotate(err, "cannot read db header")
	}
	db.computePositions()
	if err = db.readIPv4Indexes(); err != nil {
		return nil, errors.Annotate(err, "cannot read db index")
	}
	return db, nil
}

// Type gets the db type id
func (db *DB) Type() DbType {
	return db.header.Type
}

// TypeName gets the db type name
func (db *DB) TypeName() string {
	switch db.header.Type {
	case PX1:
		return "PX1"
	case PX2:
		return "PX2"
	case PX3:
		return "PX3"
	case PX4:
		return "PX4"
	default:
		return "N/A"
	}
}

// Count returns the number of records in database
func (db *DB) Count() uint32 {
	return db.header.Count
}

// Date returns the date of the current db version
func (db *DB) Date() time.Time {
	return time.Date(
		int(db.header.Year),
		time.Month(db.header.Month),
		int(db.header.Day),
		0,
		0,
		0,
		0,
		time.Local,
	)
}

// Version returns the current db version name
func (db *DB) Version() string {
	return fmt.Sprintf("%s-%d-%0.2d-%0.2d", db.TypeName(), db.header.Year, db.header.Month, db.header.Day)
}

// LookupIPV4 lookups a net.IP ipv4 address in database
func (db *DB) LookupIPV4(ip net.IP) (*Result, error) {
	ipnum, err := ipV4ToInt(ip)
	if err != nil {
		return nil, err
	}
	return db.lookupIPV4(ipnum)
}

// LookupIPV4Dot lookups a dot notation (1.2.3.4) ipv4 address in database
func (db *DB) LookupIPV4Dot(ip string) (*Result, error) {
	ipnum, err := ipV4Dot2int(ip)
	if err != nil {
		return nil, err
	}
	return db.lookupIPV4(ipnum)
}

// LookupIPV4Num lookups a numeric  ipv4 address in database
func (db *DB) LookupIPV4Num(ip uint32) (*Result, error) {
	return db.lookupIPV4(ip)
}

// parses db file header
func (db *DB) readHeader() error {
	var err error
	db.header = &dbHeader{}
	t, err := db.readUint8(0)
	if err != nil {
		return err
	}
	switch t {
	case uint8(PX1), uint8(PX2), uint8(PX3), uint8(PX4):
		db.header.Type = DbType(t)
	default:
		db.header.Type = UnknownDbType
	}
	if db.header.Type == UnknownDbType {
		return fmt.Errorf("invalid db format or unknown db type")
	}
	if err = db.readHeaderDate(); err != nil {
		return err
	}
	if err = db.readHeaderCounts(); err != nil {
		return err
	}
	if err = db.readHeaderAddrs(); err != nil {
		return err
	}
	return nil
}

// parses date in db file header
func (db *DB) readHeaderDate() error {
	year, err := db.readUint8(2)
	if err != nil {
		return err
	}
	db.header.Year = 2000 + uint16(year)
	db.header.Month, err = db.readUint8(3)
	if err != nil {
		return err
	}
	db.header.Day, err = db.readUint8(4)
	return err
}

// parses counts in db file header
func (db *DB) readHeaderCounts() error {
	var err error
	db.header.Cols, err = db.readUint8(1)
	if err != nil {
		return err
	}
	if db.header.Cols <= 0 {
		return fmt.Errorf("invalid db format")
	}
	db.header.Count, err = db.readUint32(5)
	if err != nil {
		return err
	}
	if db.header.Count <= 1 {
		return fmt.Errorf("invalid db format")
	}
	db.header.IPv4ColumnSize = db.header.Cols << 2
	return nil
}

// parses addrs in db file header
func (db *DB) readHeaderAddrs() error {
	var err error
	db.header.BaseAddr, err = db.readUint32(9)
	if err != nil {
		return err
	}
	db.header.IndexBaseAddr, err = db.readUint32(21)
	return err
}

// compute field positions according to type
func (db *DB) computePositions() {
	db.positions = &positions{}
	if countryPos[db.header.Type] != 0 {
		db.positions.Country = (countryPos[db.header.Type] - 1) << 2
	}
	if regionPos[db.header.Type] != 0 {
		db.positions.Region = (regionPos[db.header.Type] - 1) << 2
	}
	if cityPos[db.header.Type] != 0 {
		db.positions.City = (cityPos[db.header.Type] - 1) << 2
	}
	if ispPos[db.header.Type] != 0 {
		db.positions.ISP = (ispPos[db.header.Type] - 1) << 2
	}
	if proxytypePos[db.header.Type] != 0 {
		db.positions.Proxy = (proxytypePos[db.header.Type] - 1) << 2
	}
}

// read and store all ipv4 indexes
func (db *DB) readIPv4Indexes() error {
	pos := db.header.IndexBaseAddr
	for i := 0; i < maxIndexes; i++ {
		start, err := db.readUint32(pos - 1)
		if err != nil {
			return err
		}
		end, err := db.readUint32(pos + 3)
		if err != nil {
			return err
		}
		db.ipv4Indexes[i][0] = start
		db.ipv4Indexes[i][1] = end
		pos += 8
	}
	return nil
}

// lookups a record in db for an ipv4 addr
func (db *DB) lookupIPV4(ip uint32) (*Result, error) {
	pos, err := db.findPosForIPV4(ip)
	if err != nil {
		return nil, err
	}
	if pos == 0 {
		return nil, nil
	}
	res, err := db.readIPV4Record(pos + 1)
	if err != nil {
		return nil, err
	}
	res.IP = intToIPV4(ip)
	return res, nil
}

// lookups a pos in db for an ipv4 addr
func (db *DB) findPosForIPV4(ip uint32) (uint32, error) {
	indexaddr := ip >> 16
	low := db.ipv4Indexes[indexaddr][0]
	high := db.ipv4Indexes[indexaddr][1]
	for low <= high {
		mid := (low + high) / 2
		rowOffset := db.header.BaseAddr + (mid * uint32(db.header.IPv4ColumnSize)) - 1
		ipFrom, err := db.readUint32(rowOffset)
		if err != nil {
			return 0, errors.Annotate(err, "cannot read db index")
		}
		ipTo, err := db.readUint32(rowOffset + uint32(db.header.IPv4ColumnSize))
		if err != nil {
			return 0, errors.Annotate(err, "cannot read db index")
		}
		if ipFrom <= ip && ipTo >= ip {
			return rowOffset, nil
		}
		if ipFrom > ip {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return 0, nil
}

// gets the byte offset for a field
func (db *DB) getIPV4ByteOffset(field string, baseOffset uint32) uint32 {
	var idx uint8

	switch field {
	case "proxy":
		idx = (proxytypePos[db.header.Type] - 1) << 2
	case "country":
		idx = (countryPos[db.header.Type] - 1) << 2
	case "region":
		idx = (regionPos[db.header.Type] - 1) << 2
	case "city":
		idx = (cityPos[db.header.Type] - 1) << 2
	case "isp":
		idx = (ispPos[db.header.Type] - 1) << 2
	default:
		return 0
	}
	return baseOffset + uint32(idx)
}

// reads the Proxy field for record
func (db *DB) readRecordProxy(res *Result, off uint32) error {
	if db.positions.Proxy != 0 {
		addr, err := db.readUint32(db.getIPV4ByteOffset("proxy", off) - 1)
		if err != nil {
			return err
		}
		b, err := db.readStr(addr)
		if err != nil {
			return err
		}
		res.Proxy = proxyNameToProxyType(b)
		return nil
	}
	res.Proxy = ProxyNA
	return nil
}

// reads the Country field for record
func (db *DB) readRecordCountry(res *Result, off uint32) error {
	pos, err := db.readUint32(db.getIPV4ByteOffset("country", off) - 1)
	if err != nil {
		return err
	}
	short, err := db.readStr(pos)
	if err != nil {
		return err
	}
	long, err := db.readStr(pos + 3)
	if err != nil {
		return err
	}
	if short != "" && short != "-" {
		res.CountryCode = &short
	}
	if long != "" && long != "-" {
		res.Country = &long
	}
	return nil
}

// reads the Region field for record
func (db *DB) readRecordRegion(res *Result, off uint32) error {
	pos, err := db.readUint32(db.getIPV4ByteOffset("region", off) - 1)
	if err != nil {
		return err
	}
	region, err := db.readStr(pos)
	if err != nil {
		return err
	}
	if region != "" && region != "-" {
		res.Region = &region
	}
	return nil
}

// reads the City field for record
func (db *DB) readRecordCity(res *Result, off uint32) error {
	pos, err := db.readUint32(db.getIPV4ByteOffset("city", off) - 1)
	if err != nil {
		return err
	}
	city, err := db.readStr(pos)
	if err != nil {
		return err
	}
	if city != "" && city != "-" {
		res.City = &city
	}
	return nil
}

// reads the ISP field for record
func (db *DB) readRecordISP(res *Result, off uint32) error {
	pos, err := db.readUint32(db.getIPV4ByteOffset("isp", off) - 1)
	if err != nil {
		return err
	}
	isp, err := db.readStr(pos)
	if err != nil {
		return err
	}
	if isp != "" && isp != "-" {
		res.ISP = &isp
	}
	return nil
}

// reads a record
func (db *DB) readIPV4Record(off uint32) (*Result, error) {
	r := &Result{}
	if err := db.readRecordProxy(r, off); err != nil {
		return nil, err
	}
	if err := db.readRecordCountry(r, off); err != nil {
		return nil, err
	}
	if err := db.readRecordRegion(r, off); err != nil {
		return nil, err
	}
	if err := db.readRecordCity(r, off); err != nil {
		return nil, err
	}
	if err := db.readRecordISP(r, off); err != nil {
		return nil, err
	}
	return r, nil
}

// reads a uint8 value at position in file
func (db *DB) readUint8(pos uint32) (uint8, error) {
	if pos > db.dataSize-1 {
		return 0, io.EOF
	}
	return db.data[pos], nil
}

/*
// reads a uint16 value at position in file
func (db *DB) readUint16(pos uint32) (uint16, error) {
	if pos > db.dataSize - 2 {
		return 0, io.EOF
	}
	bin := db.data[pos : pos + 2]
	return fileEndianness.Uint16(bin), nil
}
*/

// reads a uint32 value at position in file
func (db *DB) readUint32(pos uint32) (uint32, error) {
	if pos > db.dataSize-4 {
		return 0, io.EOF
	}
	bin := db.data[pos : pos+4]
	return fileEndianness.Uint32(bin), nil
}

// reads a byte slice at position in file
func (db *DB) readByteSlice(pos uint32) ([]byte, error) {
	if pos > db.dataSize-1 {
		return nil, io.EOF
	}
	size, err := db.readUint8(pos)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return nil, nil
	}
	if pos+uint32(size) > db.dataSize {
		return nil, io.EOF
	}
	b := make([]byte, size)
	for i := uint8(0); i < size; i++ {
		b[i] = db.data[pos+uint32(1+i)]
	}
	return b, nil
}

// reads a string at position in file
func (db *DB) readStr(pos uint32) (string, error) {
	b, err := db.readByteSlice(pos)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// string ip to unsigned 32 bit number
func ipV4ToInt(ip net.IP) (uint32, error) {
	if ip == nil {
		return 0, fmt.Errorf("invalid IP")
	}
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16]), nil
	}
	return binary.BigEndian.Uint32(ip), nil
}

// string ip to unsigned 32 bit number
func ipV4Dot2int(ipStr string) (uint32, error) {
	return ipV4ToInt(net.ParseIP(ipStr))
}

// unsigned 32 bit number to ipv4 string
func intToIPV4(num uint32) string {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, num)
	return ip.String()
}
