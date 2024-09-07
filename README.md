# WIP CCA Selection System

There are various issues with Schoolsbuddy, the current system our school uses
for CCA registrations:

* There are various race conditions and locking issues that occur when too many
  students use the system at the same time. Course member caps are enforced by
  kicking random users out of the system, instead of having the course choice
  fail gracefully. The number of remaining seats in each course is also not
  updated live.
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
  to the year group's requirements.
* The school has to pay Schoolsbuddy a subscription fee, while something as
  simple as CCA selection could be easily accomplished on cheap hardware and
  minimal software.

There are a few things that I need to know before I could start implementing:

* How does the CCA staff enter available courses, teacher information, member
  caps, and timing information into the system? Would it be okay if I just
  accept a CSV spreadsheet?
* How do they get student choices out of the system and into PowerSchool?
* Does the school use Schoolsbuddy's payment system? I am unable to implement
  payment systems.
