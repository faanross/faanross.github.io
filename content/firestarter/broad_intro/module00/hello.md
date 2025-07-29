---
showTableOfContents: true
title: "hello."
type: "page"
---

**I'll write a longer, more comprehensive intro once the course is done, but for those who would like to start now, 
I do want to provide a brief overview.**

The idea of this course is to provide an overview of many topics you'll encounter in maldev. In my own experience, 
I found that most curricula focus on immediately going deep into a specific topic – which is of course great, 
given that this is a very technical field, and thus you need to go deep to truly get something.

But I also found that for me personally, I often felt confused about how different topics, techniques, ideas, etc., 
conceptually related to one another in this field as whole. This gave me the sense of feeling lost, like I was starting
to understand something specific, but lost in a land that was still a great mystery to me.

And so the idea here is that rather than going deep on any one specific topic, I want to go broad and introduce you to 
many different concepts in this awesome field.


Also, as far as is possible, the aim is not just to cobble together a bunch of disparate little crash courses, but rather to 
have a thread of continuity throughout. So, we'll continue to build on the same project throughout the course as we learn. 
We'll start with something crude, realize its weaknesses, improve on it, and iterate. 

And, to the extent that it's possible, I have attempted to recapitulate the evolution of maldev as it's played out over its short history.
This means that transitions largely follow the history of maldev, instead of just jumping to what are today considered 'the best techniques'. 
The reason for this is I want us to get a clear understanding of all the fundamentals, and of the collective (and reciprocal) process of 
iterative design that has led us to where we are today. 

Even if, for example, using `ntdll.dll` to bypass Win32 API hooks should probably be avoided in favor of syscalls, 
I do still think there is value in learning what it is, why it (at the time) represented the next logical step following the move away 
from the Win32 API, and how it also eventually became ineffective. Giving you this broad understanding will, in my opinion, equip
you with the meta-understanding to help drive the inevitable future evolution of the field. 

I'll also cover numerous foundational topics – like DLLs, PE structures, assembly, kernel calls, etc. – but on an 'as-needed' basis. 
I'm doing this because I personally found that going into a deep dive into these technically demanding fields before 
relating them to their relevance in maldev made me feel lost. I often felt like Sisyphus attempting to memorize all the different PE headers 
and structures, and so, here, I instead just tell you about the exact ones we'll use immediately, why we use them, why they are important, etc.

Of course, gaining deep technical proficiency in these foundations is essential if you really want to commit to this path. 
But my hope for this course is that by giving you a taste and immediately highlighting how it's relevant, I can inspire you to jump 
into those deep dives with zeal. It worked for me, so my hope is that it will work for a few of you too.

For the course, we mostly use Go, but there are times when it absolutely makes sense to use C/C++, and in minor instances, ASM. 
But again, no deep dives, no exhaustive explanations – just teaching enough to get you started and give you a feel for what 
this awesome field is all about.

Again, I'll write a much more comprehensive introduction once I've completed the entire course (it should go up until Module 20, 
at least according to my current outline). In the meantime, please feel free to send me any feedback – questions are good; 
critique is equally welcome. 

I write and frame things as they make sense to me, but if you ever feel I was unclear, an explanation was lacking, 
or a transition was too abrupt, please tell me. I'm not someone who feels personally attacked by constructive critique; 
I actually value it tremendously, since I feel like someone is giving their time and effort to allow me to create something 
better so that all of us can benefit together.

So, if you wanna connect – up at the top there is a Discord icon, feel free to join us there, or hit me up at moi@faanross.com.

Live long and prosper.
Faan

---
[|TOC|]({{< ref "../moc.md" >}})
[|NEXT|]({{< ref "../module01/intro_DLLs.md" >}})