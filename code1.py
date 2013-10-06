from math import *
class User:
    def __init__(self,level,points,interests):
        self.level=level
        self.points=points  #10*2^level
        self.interests=interests
    def update(self):
        if self.points<20: self.points=20
        self.level= int(log((self.points/10.0),2))

def matching(user1,user2,totint,totlev):
    l1=user1.level
    l2=user2.level
    i1=user1.interests
    i2=user2.interests
    i3=i1&i2
    l=1-(float(abs(l1-l2))/float(totlev))
    c=float(len(i3))/float(totint)
    w1=2
    w2=1
    match=(w1*l+w2*c)/float(w1+w2)
    return match,i3

def quest(user1,user2):
    l1=user1.level
    l2=user2.level
    i1=user1.interests
    i2=user2.interests
    l=int(ceil((l1+l2)/2.0))
    i3=i1&i2
    i11=i1-i3
    i22=i2-i3
    i4=set([(i,2) for i in i3])
    i5=i4 | set([(i,1) for i in i11]) | set([(i,1) for i in i22])
    return l,i5

def modify(user1,user2,lev,sta,totlev):
    #sta is win 1, tie/skip 0, loss -1
    l1=user1.level
    l2=user2.level
    j1=(abs(lev-l1))-(lev-l1)
    j2=(abs(lev-l2))-(lev-l2)
    dicw={1:4,2:7,3:10,4:14,5:16,6:20,7:40,8:60,9:100,10:160}
    if sta==1:
        user1.points+=dicw[lev]
        user2.points+=dicw[lev]
        user1.update()
        user2.update()
    elif sta==-1:
        user1.points-=j1
        user2.points-=j2
        user1.update()
        user2.update()
