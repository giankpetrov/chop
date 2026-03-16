## 2025-03-24 - Overly Permissive Path Exception in Security Check
**Vulnerability:** The `IsSecure` function had an exception for any path starting with `/tmp`, allowing them to be world-writable. This meant any file or subdirectory deep within `/tmp` was trusted even if insecure.
**Learning:** `strings.HasPrefix` is often too blunt for path-based security exceptions. An attacker could exploit this by creating an insecure configuration file in a subdirectory of a trusted prefix.
**Prevention:** Restrict security exceptions to exact directory matches (like `/tmp` and `/var/tmp` themselves) rather than entire prefix trees. Always use `filepath.Clean` before performing path comparisons to normalize input.
