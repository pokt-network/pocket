package bls

import (
        "github.com/herumi/bls-eth-go-binary/bls"
        "log"
        "time"
)



func main() {
        // this is not thread safe, required for BLS operations
        bls.Init(bls.BLS12_381)
        bls.SetETHmode(bls.EthModeDraft07)

        // Declare string and byte versions of variables
        n := "Pokt"
        m := []byte("Pokt")
        // BLS operations
        var aggSig *bls.Sign
        var aggPub *bls.PublicKey
        //start timer for BLS throughput test
        startTime := time.Now()
        // loop for test
        for i := 0; i < 1000; i++ {
                var sec bls.SecretKey
                sec.SetByCSPRNG()
                if i == 0 {
                        aggSig = sec.SignByte(m)
                        aggPub = sec.GetPublicKey()
                } else {
                        aggSig.Add(sec.SignByte(m))
                        aggPub.Add(sec.GetPublicKey())
                }
        }
        endTime := time.Now()
        //Inform the user
        log.Printf("Time required to sign 1000 messages and aggregate 1000 pub keys and signatures: %f seconds", endTime.Sub(startTime).Seconds())
        log.Printf("Aggregate Signature: 0x%x", aggSig.Serialize())
        log.Printf("Aggregate Public Key: 0x%x", aggPub.Serialize())

        startTime = time.Now()
        // Validate Aggregation Signing actually works, exit if not
        if !aggSig.Verify(aggPub, n) {
                log.Fatal("Aggregate Signature Does Not Verify")
        }
        log.Printf("Aggregate Signature Verifies Correctly!")
        //End timer
        endTime = time.Now()
        // Print time collected, end time counter
        log.Printf("Time required to verify aggregate sig: %f seconds", endTime.Sub(startTime).Seconds())
}




