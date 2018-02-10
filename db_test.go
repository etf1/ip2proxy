package ip2proxy_test

import (
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/etf1/ip2proxy"
)

var _ = Describe("Db", func() {
	Context("when initializing", func() {
		It("should returns an error an unexistant file", func() {
			db, err := Open("/lol/idonttexists")
			Expect(db).Should(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("cannot open/read db file: open /lol/idonttexists: no such file or directory"))
		})
		It("should returns an error on a file without read permissions", func() {
			db, err := Open(filepath.Join("testdata", "forbidden"))
			Expect(db).Should(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("cannot open/read db file: testdata/forbidden is empty or not redable"))
		})
		It("should returns an error on an empty file", func() {
			db, err := Open(filepath.Join("testdata", "empty"))
			Expect(db).Should(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("cannot open/read db file: testdata/empty is empty or not redable"))
		})
		It("should returns an error on a small random file", func() {
			db, err := Open(filepath.Join("testdata", "small"))
			Expect(db).Should(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("cannot read db header: invalid db format or unknown db type"))
		})
		It("should returns an error on a big random file", func() {
			db, err := Open(filepath.Join("testdata", "random"))
			Expect(db).Should(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("cannot read db header: invalid db format or unknown db type"))
		})
		It("should returns a valid db instance on a valid file", func() {
			db, err := Open(filepath.Join("testdata", "IP2PROXY-LITE-PX4.BIN"))
			Expect(err).To(BeNil())
			Expect(db).ToNot(BeNil())
		})
	})

	Context("when correctly initialized", func() {
		db, err := Open(filepath.Join("testdata", "IP2PROXY-LITE-PX4.BIN"))
		if err != nil {
			Fail("Loading IP2PROXY-LITE-PX4.BIN should not have failed", 1)
		}
		It("should return the valid Type infos", func() {
			Expect(db.Type()).To(Equal(PX4))
			Expect(db.TypeName()).To(Equal("PX4"))
		})
		It("should return the valid record count", func() {
			Expect(db.Count()).To(Equal(uint32(3445221)))
		})
		It("should return the valid version string", func() {
			Expect(db.Version()).To(Equal("PX4-2018-02-01"))
		})
		It("should return the valid db creation date", func() {
			Expect(db.Date()).To(Equal(time.Date(2018, time.Month(2), 1, 0, 0, 0, 0, time.Local)))
		})
	})
	Context("when looking up", func() {
		db, err := Open(filepath.Join("testdata", "IP2PROXY-LITE-PX4.BIN"))
		if err != nil {
			Fail("Loading IP2PROXY-LITE-PX4.BIN should not have failed", 1)
		}
		It("should return an error for invalid ips", func() {
			res, err := db.LookupIPV4Dot("289.1.2.3")
			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid IP"))
			res, err = db.LookupIPV4(nil)
			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid IP"))
		})
		It("should return a valid info for proxy hosts", func() {
			list := map[string]ProxyType{
				"78.220.10.108": ProxyNOT,
				"8.8.8.8":       ProxyDCH,
				"1.0.132.186":   ProxyPUB,
				"1.32.122.154":  ProxyWEB,
				"2.7.154.188":   ProxyTOR,
				"1.0.194.42":    ProxyVPN,
			}
			for ip, expected := range list {
				res, err := db.LookupIPV4Dot(ip)
				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
				Expect(res.Proxy).To(Equal(expected))
			}
		})
		It("should return a valid info in the country field", func() {
			ptrStr := func(str string) *string { return &str }
			list := map[string]*string{
				"78.220.10.108":   nil,
				"217.212.231.208": ptrStr("Poland"),
				"2.6.120.66":      ptrStr("France"),
			}
			for ip, expected := range list {
				res, err := db.LookupIPV4Dot(ip)
				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
				Expect(res.Country).To(Equal(expected))
			}
		})
		It("should return a valid info in the ISP field", func() {
			ptrStr := func(str string) *string { return &str }
			list := map[string]*string{
				"78.220.10.108":   nil,
				"217.212.231.208": ptrStr("Opera Software ASA"),
				"2.6.120.66":      ptrStr("France Telecom S.A."),
			}
			for ip, expected := range list {
				res, err := db.LookupIPV4Dot(ip)
				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
				Expect(res.ISP).To(Equal(expected))
			}
		})
		It("should return a valid info in the City field", func() {
			ptrStr := func(str string) *string { return &str }
			list := map[string]*string{
				"78.220.10.108":   nil,
				"212.9.226.39":    ptrStr("Kiev"),
				"206.190.140.157": ptrStr("Providence"),
				"74.219.56.231":   ptrStr("Columbus"),
			}
			for ip, expected := range list {
				res, err := db.LookupIPV4Dot(ip)
				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
				Expect(res.City).To(Equal(expected))
			}
		})
		It("should return a valid info in the City field", func() {
			ptrStr := func(str string) *string { return &str }
			list := map[string]*string{
				"78.220.10.108": nil,
				"186.94.238.11": ptrStr("Trujillo"),
				"207.224.64.81": ptrStr("Minnesota"),
				"46.151.249.26": ptrStr("Chernivetska oblast"),
			}
			for ip, expected := range list {
				res, err := db.LookupIPV4Dot(ip)
				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
				Expect(res.Region).To(Equal(expected))
			}
		})
	})
})
