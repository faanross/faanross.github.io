package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/miekg/dns"
)

// our local dns server's port
var port = 53530

// The zone map will store our DNS records.
var zone = map[string][]dns.RR{}

// setupZone initializes our DNS records for our domain (timeserversync.com)
func setupZone() {
	domain := "timeserversync.com." // Note the trailing dot to make it an FQDN.

	// A Record (IPv4 Address)
	// Maps www.timeserversync.com to an IPv4 address.
	aRecord, _ := dns.NewRR(fmt.Sprintf("%s A 192.0.2.1", "www."+domain))
	zone["www."+domain] = append(zone["www."+domain], aRecord)

	// AAAA Record (IPv6 Address)
	// Maps api.timeserversync.com to an IPv6 address.
	aaaaRecord, _ := dns.NewRR(fmt.Sprintf("%s AAAA 2001:db8::1", "api."+domain))
	zone["api."+domain] = append(zone["api."+domain], aaaaRecord)

	// MX Record (Mail Exchange)
	// Specifies the mail server for the domain.
	// Priority 10, target mail.timeserversync.com.
	mxRecord, _ := dns.NewRR(fmt.Sprintf("%s MX 10 mail.%s", domain, domain))
	zone[domain] = append(zone[domain], mxRecord)
	// We also need an A record for the mail server itself.
	mailARecord, _ := dns.NewRR(fmt.Sprintf("%s A 192.0.2.2", "mail."+domain))
	zone["mail."+domain] = append(zone["mail."+domain], mailARecord)

	// SOA Record (Start of Authority)
	// This is a crucial record for any authoritative server. It provides administrative
	// information about the zone.
	soaRecord, _ := dns.NewRR(fmt.Sprintf("%s SOA ns1.%s hostmaster.%s 2025071501 7200 3600 1209600 3600",
		domain, domain, domain))
	zone[domain] = append(zone[domain], soaRecord)
	// We also need an A record for the nameserver mentioned in the SOA record.
	ns1ARecord, _ := dns.NewRR(fmt.Sprintf("%s A 192.0.2.3", "ns1."+domain))
	zone["ns1."+domain] = append(zone["ns1."+domain], ns1ARecord)

	log.Println("DNS zone for timeserversync.com setup complete.")
}

// handleDNSRequest is the core logic for our server. It gets called for each incoming query.
func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// Create a response message.
	m := new(dns.Msg)
	// Use SetReply to copy the request's header, including the transaction ID.
	// This is crucial for the client to match the response to its original query.
	m.SetReply(r)
	m.Compress = false // Disable compression for simplicity.

	// The 'aa' (Authoritative Answer) flag is the most important part of our response.
	// It tells the client that we are the definitive source for this domain.
	m.Authoritative = true

	// We only have one question in a standard query.
	q := r.Question[0]
	log.Printf("Received query for: %s, Type: %s", q.Name, dns.TypeToString[q.Qtype])

	// Check if the query is for our authoritative domain.
	// We use strings.HasSuffix to check if the query name ends with our domain name.
	if !strings.HasSuffix(strings.ToLower(q.Name), "timeserversync.com.") {
		// If the query is for a domain we are not authoritative for, we refuse it.
		m.Rcode = dns.RcodeRefused
		log.Printf("Refusing query for non-authoritative domain: %s", q.Name)
	} else {
		// Look for records matching the query name and type.
		records, found := zone[strings.ToLower(q.Name)]
		if found {
			for _, rec := range records {
				// Check if the record type matches the query type.
				if rec.Header().Rrtype == q.Qtype {
					m.Answer = append(m.Answer, rec)
				}
			}
		}

		// If no records were found in the Answer section, it means either the specific
		// subdomain doesn't exist or the record type doesn't exist.
		// A proper authoritative server should return the SOA record in the Authority section
		// to indicate it's authoritative but has no answer.
		if len(m.Answer) == 0 {
			// Find the SOA record for the base domain to include in the response.
			soa, ok := zone["timeserversync.com."]
			if ok {
				// We add the SOA record to the Authority section (Ns).
				m.Ns = append(m.Ns, soa[0])
			}
		}
	}

	// Send the response back to the client.
	err := w.WriteMsg(m)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func main() {
	// Step 1: Initialize our zone records.
	setupZone()

	// Step 2: Attach the handler function to the DNS server.
	// We tell the server to call `handleDNSRequest` for any query.
	dns.HandleFunc(".", handleDNSRequest)

	// Step 3: Start the server.

	server := &dns.Server{Addr: fmt.Sprintf(":%d", port), Net: "udp"}
	log.Printf("Starting authoritative DNS server on port %d\n", port)

	// The ListenAndServe method blocks, so the program will stay running.
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n", err.Error())
	}
	defer server.Shutdown()
}
