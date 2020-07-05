$TTL 1M
;1D
@			IN SOA		ns01.example.com.	hostmaster.example.com. (
			2020053001	; serial
			3H		; refresh
			1H		; retry
			7D		; expire
			1M)		; negcache TTL
			IN NS		ns01.example.com.
			IN NS		ns02.example.com.

test			IN A		192.0.2.1
;VFFI7ZOMWGN2MHCMBBZ5HEPJQ6MC7O6T	IN TXT foo
