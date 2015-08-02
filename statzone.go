/*****************************************************************************/
/*                                                                           */
/* StatZone (c) by Frederic Cambus 2012-2015                                 */
/* http://www.statdns.com                                                    */
/*                                                                           */
/* Created: 2012/02/13                                                       */
/* Last Updated: 2015/08/02                                                  */
/*                                                                           */
/* StatZone is released under the BSD 3-Clause license.                      */
/* See LICENSE file for details.                                             */
/*                                                                           */
/*****************************************************************************/

package main

import (
	"bufio"
	"fmt"
	"github.com/miekg/dns"
	"os"
	"strings"
)

type Domains struct {
	count    int
	idn      int
	previous string
	suffix   int
}

var rrParsed int

// Return rdata
func rdata(RR dns.RR) string {
	return strings.Replace(RR.String(), RR.Header().String(), "", -1)
}

func main() {
	header := `-------------------------------------------------------------------------------
                   StatZone (c) by Frederic Cambus 2012-2015
-------------------------------------------------------------------------------`

	inputFile := os.Args[1]

	fmt.Println(header)

	fmt.Println("\nParsing zone :", inputFile)

	domains := new(Domains)

	ns := map[string]int{}
	signed := map[string]int{}

	zoneFile, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("ERROR : Can't open zone file.")
	}

	zone := dns.ParseZone(bufio.NewReader(zoneFile), "", "")

	var rrtypes [100]int

	for parsedLine := range zone {
		if parsedLine.RR != nil {
			rrtypes[parsedLine.RR.Header().Rrtype]++

			switch parsedLine.RR.Header().Rrtype {
			case dns.TypeDS:
				/* Increment Signed Domains counter */
				signed[parsedLine.RR.Header().Name]++
			case dns.TypeNS:
				/* Increment NS counter */
				ns[rdata(parsedLine)]++

				if parsedLine.RR.Header().Name != domains.previous { // Unique domain

					/* Increment Domain counter */
					domains.count++
					domains.previous = parsedLine.RR.Header().Name

					/* Check if the domain is an IDN */

					if strings.HasPrefix(strings.ToLower(parsedLine.RR.Header().Name), "xn--") {
						domains.idn++
					}

					/* Display progression */
					if domains.count%1000000 == 0 {
						fmt.Printf("*")
					} else if domains.count%100000 == 0 {
						fmt.Printf(".")
					}
				}
			}
		} else {
			fmt.Println("ERROR : A problem occured while parsing the zone file.")
		}

		/* Increment number of resource records parsed */
		rrParsed++
	}

	/* Don't count origin */
	domains.count--

	fmt.Println("\n---[ Parsing results ]---------------------------------------------------------\n")
	fmt.Println(rrParsed, "RRs parsed.")
	for loop := 0; loop < len(rrtypes); loop++ {
		rrtype := rrtypes[loop]
		if rrtype != 0 {
			fmt.Println(dns.TypeToString[uint16(loop)], "records :", rrtype)
		}
	}

	fmt.Println("\n---[ Results ]-----------------------------------------------------------------\n")
	fmt.Println("Domains : ", domains.count)
	fmt.Println("DNSSEC Signed : ", len(signed))

	fmt.Println("IDNs : ", domains.idn)
	fmt.Println("NS : ", len(ns))

	fmt.Println("\n---[ CSV values ]--------------------------------------------------------------\n")

	fmt.Println("IPv4 Glue ; IPv6 Glue ; NS ; Unique NS ; DS ; Signed ; IDNs ; Domains")
	fmt.Println(rrtypes[dns.TypeA], ";", rrtypes[dns.TypeAAAA], ";", rrtypes[dns.TypeNS], ";", len(ns), ";", rrtypes[dns.TypeDS], ";", len(signed), ";", domains.idn, ";", domains.count)
}
