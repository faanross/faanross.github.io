---
showTableOfContents: true
title: "Key Derivation Logic(Theory 7.2)"
type: "page"
---
## Key Derivation: Avoiding Hardcoded Secrets

We've seen how rolling XOR can help to improve the obfuscation of the data, but the effectiveness of this method hinges entirely on the key. In truth, rolling XOR will make reversing more challenging but not impossible, but if a key is easily found, then it no longer presents much of a challenge at all.

So simply embedding a fixed, hardcoded key string directly in the loader or agent program (like `mySecretKey = "password123"`) is poor practice for anything intended to be stealthy. Such keys can often be found relatively easily using static analysis tools (like string searching) or by observing the deobfuscation routine during dynamic analysis.

## So, What is Key Derivation?

Instead of directly writing the final secret key into the code (hardcoding), **key derivation is the process of generating the required key dynamically when it's needed, based on some other initial information.**

Think of it this way:
**Hardcoding:** You write the final answer (`key = 12345`) directly into your program. Anyone looking at the code can easily find `12345`.

**Key Derivation:** You store some starting values (e.g., `valueA = 100`, `valueB = 23`) and a _process_ or formula (e.g., `key = (valueA * 100) + (valueB * 100) + 45`). The program _calculates_ `key = 12345` only when it needs to use it.

The "initial information" used for derivation could be:

- A master password or passphrase.
- A pre-shared secret value (which itself might be embedded or disguised).
- Non-secret data (like a username, timestamp, or system ID) combined with a secret value.
- Data derived from the program's environment or configuration.

So there are different approaches - sometimes the intent is to make this initial info has hard as possible to find, sometimes it's about "hiding in plain sight", and sometimes it's about combining the two. **There's no "fixed" approach here, there are conceptually an infinite amount of ways to derive a key, but the goal is always the same - to make it as hard as possible to figure out what it is.**

## Key Derivation Function (KDF)

In addition to the initial information, one also need to decide on a specific algorithm, called a **Key Derivation Function (KDF)**, in order to derive the final key. KDFs are designed to take the initial information and securely transform it into one or more cryptographic keys of a desired length and format.

The core idea is that the final, sensitive key used for the actual encryption/obfuscation (like the `baseKey` for our rolling XOR) doesn't exist explicitly in the stored program code or configuration files. Instead, the code contains the _components_ (initial info) and the _logic_ (KDF) needed to construct the key at runtime.

Note that whereas I mentioned that as far as our initial information goes there are conceptually an infinite amount of ways this can be created/derived, with KDFs we typically use stablished cryptographic algorithms like PBKDF2, HKDF, Argon2 etc.

Now that we understand that key derivation means _calculating_ the key instead of storing it directly, let's explore the advantages of doing this.

## Why Use Key Derivation?

The primary goals of using key derivation techniques instead of hardcoded keys are:

1. **Improve Stealth:** Avoid having obvious key material sitting in the compiled binary. The key should be generated only when needed.
2. **Increase Analysis Cost:** Make it harder for an analyst to quickly identify and extract the key required to deobfuscate payloads or communications. The analyst now has to understand the _derivation process_ itself. As I said above, it's not about making it impossible, but increasing the cost to the level where it might not be worth an analyst's time.
3. **Limit Impact of Compromise:** If the key changes frequently (e.g., derived using dynamic data like timestamps), the compromise of information needed to derive one key might only allow decryption of data associated with that specific key, not all past or future data.


## An Example Key Derivation Method

Let's end this section with an example of a key derivation method I've used a number of times, we'll do something similar in our labs. It employs a two-stage approach to derive the final key used for  rolling XOR obfuscation.


### Stage 1: "Shared Secret" Generation

This stage aims to create a reproducible "secret" value shared between the client and server, but embedded in a way that doesn't immediately look like a key. In other words, it takes the "hiding in plain sight" approach, at least initially, and then combines those values to produce the secret.

So for example, since a reflective loader contains a long list of constants with many generic sounding names, I add a few more that appear, at least on the surface, to be valid PE constants. As an example - `SECTION_ALIGN_REQUIRED`, `FILE_ALIGN_MINIMAL`, `PE_CHECKSUM_SEED`, etc. On quick observation these would not seem out of place, and so for a human analyst just doing an initial visual sweep over strings extracted from the binary these might be missed.

These constants had assigned values that, when combined in the correct order, would produce a secret. Of course you don't just need combine them - as I said above, you can let your imagination run wild here, while keeping in mind that the logic to reassemble the secrets will also have to exist in the application, so something hopelessly contrived might itself stand out. The goal with evasion is almost always to do things in a way that blends in the most with what appears to be "normal", or legitimate.

Also note that the function used to create the shared secret from these values should, at the very least, have names that help them blend in, let's say `getPESectionAlignmentString` and `verifyPEChecksumValue` as simple examples.

There will need to be in final secret generating function tasked with calling the various helper functions to produce the final secret key. So again, functions combined here should "make sense", and if you have 24 functions doing extremely odd arithmetic acrobatics you've probably just wasted your time. I recommend looking at "legitimate" code that interacts with PE files, and using that as seeds of inspiration on how you can emulate a function that appears to be benign, yet derives the secret.

Also keep in mind - this secret needs to be produced not only on the agent's side, but also on the server side since it will be tasked with encrypting the data, and the key obviously needs to be the same for this to work.

But once you have this secret, it does not end there.


### Stage 2: Session Key Derivation

The shared secret generated in Stage 1 provides a base, but using the same key for every communication would still be weak since discovering it once essentially burns it, rendering it useless for all future operations. So the goal of Stage 2 is to ensure the ultimate key that is derived from the secret is unique by incorporating some dynamic information.

Here's a conceptual example.

The agent can include dynamic data in its request to the server, specifically the current Unix timestamp and a generated Client ID (derived from system information), embedded within the User-Agent string. The server extracts this timestamp and Client ID from the User-Agent.

There is then a function, let's call it `deriveKeyFromParams` (again present on both sides of the equation) takes the shared secret (from Stage 1), the timestamp, and the client ID, and concatenates them into a single string (`combined := sharedSecret + timestamp + clientID`).

The code then performs a very basic key derivation function: it creates a fixed-size key buffer (e.g., 32 bytes) and populates it by repeating the `combined` string until the buffer is full. This resulting byte slice becomes the final `baseKey` used in the rolling XOR function (`obfuscatePayload` on the server, `deobfuscatePayload` on the agent).

So the key itself is not fixed/static, nor is it ever sent over the wire - it's a product that both side are able to calculate independently.

Even if an analyst intercepts one payload and its corresponding key derivation parameters (timestamp, client ID), deriving the key wouldn't necessarily help them decrypt payloads from different sessions or different clients, because the dynamic inputs would be different. The underlying shared secret derived from constants remains hidden within the binary logic, making it harder to find than a simple hardcoded string.

## Conclusion
While the specific key derivation function shared here is still relatively simple and not cryptographically strong by modern standards (a proper key derivation function like PBKDF2 or HKDF would be more secure), it demonstrates the principle of combining a shared, embedded secret with dynamic data to generate session-specific keys, significantly improving upon hardcoded or static keys for obfuscation.

Let's now implement our newfound knowledge of rolling XOR and this key derivation logic in our reflective loader.





---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "rolling.md" >}})
[|NEXT|]({{< ref "rolling_lab.md" >}})