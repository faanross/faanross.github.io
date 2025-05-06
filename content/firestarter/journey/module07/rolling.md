---
showTableOfContents: true
title: "Rolling XOR (Theory 7.1)"
type: "page"
---

## Improving Obfuscation

While simple XOR provides a basic level of obfuscation, its predictability, especially with repeating keys, makes it vulnerable to analysis. Even without a key, it is extremely vulnerable to frequency analysis. To overcome some of these weaknesses, we can employ techniques to make the XOR process less static. One common improvement is known as **Rolling XOR**.

## Rolling XOR Concept

The core idea behind Rolling XOR (also sometimes called a keystream cipher based on XOR) is to avoid applying the _same_ repeating key byte sequence directly to the plaintext. Instead, the key byte used for the XOR operation at each position is _modified_ based on its position within the data stream (or other changing factors).

This means that even if the same plaintext byte appears multiple times, it will likely be XORed with a _different_ key byte each time, resulting in different ciphertext bytes. Similarly, identical sequences in the plaintext are less likely to produce identical sequences in the ciphertext. It's worth noting that though this will obviously improve defence against frequency analysis, it will produce higher levels of entropy. So as always with obfuscation, one needs to strike a balance - too much of a good thing turns out to be bad in a different way.

## Purpose: Thwarting Frequency Analysis

The primary goal of rolling XOR is to make **frequency analysis** significantly harder compared to simple XOR with a repeating key. In simple XOR, if the key `K` repeats every 4 bytes, then byte 0 of the plaintext is always XORed with `K[0]`, byte 1 with `K[1]`, ..., byte 4 with `K[0]`, byte 5 with `K[1]`, and so on. An attacker could potentially analyze every 4th byte of the ciphertext separately to deduce the corresponding key byte based on expected plaintext frequencies.

Rolling XOR disrupts this direct relationship. Since the effective key byte changes at each position `i`, the statistical patterns of the plaintext are smeared across the ciphertext, making it much harder to deduce the underlying base key or patterns simply by looking at byte frequencies in the ciphertext.

## Example Implementation

One common way to implement rolling XOR is to combine the base key byte with the current position index before performing the XOR operation.

Consider the following function:

```Go
// Pseudocode based on the project files
func obfuscatePayload(data []byte, baseKey []byte) []byte {
    keyLen := len(baseKey)
    result := make([]byte, len(data))

    for i := 0; i < len(data); i++ {
        // 1. Get the original key byte for this position
        originalKeyByte := baseKey[i % keyLen]

        // 2. Get a value based on the position (e.g., lower 8 bits of index)
        positionByte := byte(i & 0xFF) // Mask 'i' to get a single byte value

        // 3. Calculate the 'rolling' key byte by XORing original key and position
        rollingKeyByte := originalKeyByte ^ positionByte

        // 4. XOR the data with the *rolling* key byte
        result[i] = data[i] ^ rollingKeyByte
    }
    return result
}
```


- `baseKey[i % keyLen]` provides the repeating base key byte.
- `byte(i & 0xFF)` creates a modifier byte derived from the current index `i`. Using `& 0xFF` ensures this modifier cycles every 256 bytes.
- The `rollingKeyByte` is calculated by XORing these two components. This ensures that even if `baseKey` repeats, the `rollingKeyByte` changes based on `i`.
- The actual data byte `data[i]` is then XORed with this unique `rollingKeyByte`.


## Example Walkthrough
To really help solidify this concept let's walk through a very elementary example of how this would play out conceptually.


**Input Data (`data`)**: Let's say our data bytes are `data = [0x41, 0x42, 0x43, 0x44]`
**Base Key (`baseKey`)**: Let's use a short key, say `baseKey = [0x1A, 0x2B]`

We'll follow the logic inside the `for` loop for each byte.

---

**Iteration 1: Processing `data[0]` (The 1st byte, `i = 0`)**

1. **Get `originalKeyByte`**:
    - Index `i` is 0. `keyLen` is 2.
    - `i % keyLen` = `0 % 2` = `0`.
    - `originalKeyByte = baseKey[0]` = `0x1A`.
2. **Get `positionByte`**:
    - Index `i` is 0.
    - `i & 0xFF` = `0 & 0xFF` = `0`. (0xFF is binary `11111111`. `0 &` anything is `0`).
    - `positionByte = byte(0)` = `0x00`.
3. **Calculate `rollingKeyByte`**:
    - `rollingKeyByte = originalKeyByte ^ positionByte`
    - `rollingKeyByte = 0x1A ^ 0x00` = `0x1A`.
    - _(Binary: `00011010 ^ 00000000 = 00011010`)_
4. **XOR Data with `rollingKeyByte`**:
    - `result[0] = data[0] ^ rollingKeyByte`
    - `result[0] = 0x41 ^ 0x1A`
    - _(Binary: `01000001 ^ 00011010 = 01011011`)_
    - `result[0] = 0x5B`.

---

**Iteration 2: Processing `data[1]` (The 2nd byte, `i = 1`)**

1. **Get `originalKeyByte`**:
    - Index `i` is 1. `keyLen` is 2.
    - `i % keyLen` = `1 % 2` = `1`.
    - `originalKeyByte = baseKey[1]` = `0x2B`.
2. **Get `positionByte`**:
    - Index `i` is 1.
    - `i & 0xFF` = `1 & 0xFF` = `1`.
    - `positionByte = byte(1)` = `0x01`.
3. **Calculate `rollingKeyByte`**:
    - `rollingKeyByte = originalKeyByte ^ positionByte`
    - `rollingKeyByte = 0x2B ^ 0x01`
    - _(Binary: `00101011 ^ 00000001 = 00101010`)_
    - `rollingKeyByte = 0x2A`.
4. **XOR Data with `rollingKeyByte`**:
    - `result[1] = data[1] ^ rollingKeyByte`
    - `result[1] = 0x42 ^ 0x2A`
    - _(Binary: `01000010 ^ 00101010 = 01101000`)_
    - `result[1] = 0x68`.

---

**Iteration 3: Processing `data[2]` (The 3rd byte, `i = 2`)**

1. **Get `originalKeyByte`**:
    - Index `i` is 2. `keyLen` is 2.
    - `i % keyLen` = `2 % 2` = `0`.
    - `originalKeyByte = baseKey[0]` = `0x1A`. (Notice the base key repeats)
2. **Get `positionByte`**:
    - Index `i` is 2.
    - `i & 0xFF` = `2 & 0xFF` = `2`.
    - `positionByte = byte(2)` = `0x02`.
3. **Calculate `rollingKeyByte`**:
    - `rollingKeyByte = originalKeyByte ^ positionByte`
    - `rollingKeyByte = 0x1A ^ 0x02`
    - _(Binary: `00011010 ^ 00000010 = 00011000`)_
    - `rollingKeyByte = 0x18`. **Crucially, even though the `originalKeyByte` (0x1A) is the same as in Iteration 1, the `rollingKeyByte` (0x18) is different because the `positionByte` changed.**
4. **XOR Data with `rollingKeyByte`**:
    - `result[2] = data[2] ^ rollingKeyByte`
    - `result[2] = 0x43 ^ 0x18`
    - _(Binary: `01000011 ^ 00011000 = 01011011`)_
    - `result[2] = 0x5B`.

---

**Iteration 4: Processing `data[3]` (The 4th byte, `i = 3`)**

1. **Get `originalKeyByte`**:
    - Index `i` is 3. `keyLen` is 2.
    - `i % keyLen` = `3 % 2` = `1`.
    - `originalKeyByte = baseKey[1]` = `0x2B`. (Base key repeats)
2. **Get `positionByte`**:
    - Index `i` is 3.
    - `i & 0xFF` = `3 & 0xFF` = `3`.
    - `positionByte = byte(3)` = `0x03`.
3. **Calculate `rollingKeyByte`**:
    - `rollingKeyByte = originalKeyByte ^ positionByte`
    - `rollingKeyByte = 0x2B ^ 0x03`
    - _(Binary: `00101011 ^ 00000011 = 00101000`)_
    - `rollingKeyByte = 0x28`. **Again, compare to Iteration 2: same `originalKeyByte` (0x2B) but different `positionByte` leads to a different `rollingKeyByte` (0x28 vs 0x2A).**
4. **XOR Data with `rollingKeyByte`**:
    - `result[3] = data[3] ^ rollingKeyByte`
    - `result[3] = 0x44 ^ 0x28`
    - _(Binary: `01000100 ^ 00101000 = 01101100`)_
    - `result[3] = 0x6C`.

---

**Summary of Results**

- Input Data: `[0x41, 0x42, 0x43, 0x44]`
- Base Key: `[0x1A, 0x2B]`
- Rolling Keys used: `[0x1A, 0x2A, 0x18, 0x28]`
- Output Result: `[0x5B, 0x68, 0x5B, 0x6C]`


As you can see, even with a short, repeating base key (`0x1A, 0x2B`), the actual key used for XORing each byte (`0x1A, 0x2A, 0x18, 0x28`) changes progressively because it incorporates the byte's position (`i`).  The `& 0xFF` ensures the position modifier cycles every 256 bytes, adding another layer to the key variation.


## Decryption

Just like simple XOR, rolling XOR implemented this way remains perfectly symmetric. The decryption process uses the _exact same logic_ to generate the `rollingKeyByte` at each position `i`.

`PlaintextByte = CiphertextByte XOR rollingKeyByte`

Where `rollingKeyByte` is calculated identically using the `baseKey` and the position `i`. Since `(P XOR K') XOR K' = P`, applying the same deterministic rolling key generation process twice recovers the original plaintext.

## Conclusion
While rolling XOR is still not cryptographically secure in a formal sense (it's a form of stream cipher susceptible to certain attacks if parts of the plaintext/key are known), it provides a significant improvement over simple XOR for basic obfuscation purposes by defeating simple frequency analysis. The next challenge, however, remains key management – how do the encrypting and decrypting sides securely agree on the `baseKey`?

In the next section we'll explore more secure key derivation concepts to generate a shared secret and derive the final XOR key.



---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../module06/load_lab.md" >}})
[|NEXT|]({{< ref "key.md" >}})