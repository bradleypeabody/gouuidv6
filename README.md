
"Version 6" UUIDs in Go
-----------------------

Motivation
==========

(modern databases need IDs with the following properties: globally unique,
easily sort in time sequence, time component extractable as a "create date"; 
unfortunately, none of the versions specified in RFC 4122 meet all of
these requirements; while it is of course possible to define a sort order
which sorts Version 1 UUIDs according to time, this requires each system
which uses the UUID to implement such sorting and at the expense of 
some performance (Cassandra does this, but many other systems do not);
however the here-mentioned "Version 6? UUID scheme addresses this problem
in the more straightforward manner of simply storing the time component in
the UUID in the most useful sequence, so that sorting the UUID based on
it's raw bytes yeilds correct sequence.)

Encoding
========

RFC 4122 


Implementation Details: Avoiding Duplicates
===========================================

In the RFC, version 1 UUIDs lack adequate provisions to avoid preventing
duplicate UUIDs being generated at the same time.  With a clock granularity of
100 nanoseconds, it is completely possible for duplicate UUIDs to be generated
if one follows the specification naively.  The document carefully indicates
provisions for the case where the clock moves backwards in time, and yet
does not address the more usual case of simultaneous UUID generation.

When you step back from the pedantry, you can conceptually think of the first
8 bytes being devoted to storing the time (and the UUID version number, which
is a constant value), and the last 8 bytes being devoted to avoiding
duplication both on the same device and among other devices.  Of this last
8 bytes, 2 bits are the variant, 14 bits are the "clock sequence", and 48 bits
are the "node".  The clock sequence is presumably there to prevent duplication
on the same device, whereas the "node" indicates which device and avoids
duplication across devices.
