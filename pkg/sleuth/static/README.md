# Static Content

This subpackage is very much a work in progress - at the time of writing, this isn't actually used for anything but the basic framework is here, so follow on PR's will be made to add the functionality.

## Background Context

```
2/18/2025

manderson:
general approach question.... lets say that you have a relatively large amount of files (upwards of 100-something files), but are derived from only a few sources (github repos, only a handful, sane enough to build a minor configuration structure around) but some of the files can be relatively large in size (20-30mb uncompressed). They change semi-frequently (some may be updated daily, most others weekly, some very infrequent). I doubt the total size of the files would exceed a few hundred MB. Thinking through "solutions".... would you use embed and ship them as a part of the binary and parse the data as a byte slice, or host the files somewhere (e.g. s3 bucket, cloudflare R2) and fetch them on startup / cache them locally but pull them dynamically every time? Was thinking of making a small CI/CD process that checks for updates of the files, opens PRs / commits them, or making some functions that pull the files / bundle them on build but regardless, using embed to ship them with a binary (just possibly a larger than normal binary) OR make it only read the files into memory regardless of where they are hosted and cache the files

8:58
alternatively use devops ninja skills and rub some nginx on it and skip the go part of it entirely ¯\_(ツ)_/¯

sfunk 9:00 AM
I would not use embed for anything that changes with any frequency especially daily; honestly the nginx is probably the most sane approach to hosting files

9:04 what kind of files are these?

manderson 9:05 AM
lists of public resolvers, badword lists, subdomain enumeration files, cname takeover fingerprints, email blacklists, etc.

it feels somewhat problematic to make fetching these a runtime dependency vs. a build dependency, though

this is probably just due to personal comfort, but i'd rather have my CI process / build + release process shipping new binaries more frequently with the files embedded than I would rely on network + uptime of where the files are hosted + pulling the contents across the wire on runtime

sfunk 9:07 AM
yeah; well the context of what the files are I'd actually agree

because if they are out of date for a few days; thats not actually a big deal


manderson
you could hypothetically hash the contents and perform a checksum on the upstream and merge the difference but that feels messy; i've always been team #shiplatest and a solid CI / CD release process and you're just shipping bigger binaries more frequently but can do so much more reliably than a service that builds in the runtime dependencies and then potentially serves incomplete data and/or has variability in the responses across many instances of the service

9:12
i thought about a 2 phase process (one that's checking / updating the file sources and then pushing that data into a database or KV or similar) and then another that reads from it... but this data being stale isn't a huge deal and that seems like a rube goldberg situation


sfunk 9:13 AM
this might be just my preferences; but I'ld build that into its own container; and make a multistage build for your final image - that way you could decouple those releases if you needed for whatever reason

manderson 9:13 AM
the code from the files ?

sfunk 9:13 AM
oh actually - you can't if you are using embded

soo nevermind

manderson 9:15 AM
the reason I was leaning towards not doing it that way is I found quite a few golang forums talking about how using embed is much more efficient than os package and if you put the files into the container you're relying on the OS / kernel layer to perform the access of the file over the files being available directly in memory; embed uses fs (and no os)

9:15
embed doesn’t use os package and it does a raw reading of a file or dir using the fs module instead, saving literally the file content into a string, executing a direct mapping of the file content on disk. At runtime you don’t have the latency of opening it using os functions, which also does a lot of tricks behind the scene (in some situations you want avoid them). So, if you want use a read-only db file or text file, configuration and so on, embed feature is really useful.
BUT, embedding a full file o dir content into a Go binary doesn’t mean it’s all in memory at runtime, because binaries are loaded by OS a bunch at a time on demand, and being part of the binary the same can happen for the embedded file if it is a bit large. This fact doesn’t depend on Go runtime or compiler.

sfunk 9:17 AM
yep ^ that makes a lot of sense
```