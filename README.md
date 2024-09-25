# YKPS CCA Selection System (WIP)

[![Woodpecker CI](https://ci.codeberg.org/api/badges/13763/status.svg)](https://ci.codeberg.org/repos/13763)

**Note:** This is my internal assessment for the IB Diploma Programme's
Computer Science (Higher Level) course.

## Configuring, building and running

1. Copy `docs/cca.scfg.example` to `cca.scfg` and edit it to taste. The
   configuration file is well-documented. You probably need to add a client
   secret, change some paths, and change some listening options.
2. `make`
3. `./cca`

## Problems with the existing system

There are various issues with Schoolsbuddy, the current system our school uses
for CCA registrations:

* There are various race conditions and locking issues that occur when too many
  students use the system at the same time. Course member caps are enforced by
  kicking random users out of the system while over-filling some courses,
  instead of having the course choice fail gracefully. The number of remaining
  seats in each course is also not updated live.
* If a lot of people are trying to load the Schoolsbuddy page, it takes ages.
* The web interface is extremely bloated and contains more than 10 megabytes of
  unnecessary JavaScript.
* The main interface displays three CCA periods, and inside each CCA period you
  can choose between Monsday/Wednesday and Tuesday/Thursday. That's a bit
  unintuitive.
* It seems to only support authentication by PowerSchool single sign-on, which
  is problematic because PowerSchool only supports a limited number of students
  logging on at the same time.
* The selection system does not enforce CCA hours requirements, and the CCA
  office's staff must manually verify that students have completed CCA choices
  to the year group's requirements, by literally printing out the spreadsheet
  to paper and reading through them.
* The school has to pay Schoolsbuddy a subscription fee, while something as
  simple as CCA selection could be easily accomplished on cheap hardware and
  minimal software.

## Key questions

* How does the CCA staff enter available courses, teacher information, member
  caps, and timing information into the system? Would it be okay if I just
  accept a CSV spreadsheet?   
  Apparently they use the Schoolsbuddy web page to do so. I'd add a CSV upload
  as that would be more convenient for most purposes.
* How do they get student choices out of the system and into PowerSchool?   
  It looks like they export an Excel sheet. See `docs/fields.txt`.
