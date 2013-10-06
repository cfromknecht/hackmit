from math import *
class User:
    def __init__(self,level,points,interests):
        self.level=level
        self.points=points  #10*(2^level) is minimum points for a particular level
        self.interests=interests #interests is a set
    def update(self):
        if self.points<20: self.points=20
        self.level= int(log((self.points/10.0),2))

def matching(user1,user2,totint,totlev):
    #this function used to give the matching between two users
    l1=user1.level
    l2=user2.level
    i1=user1.interests
    i2=user2.interests
    i3=i1&i2 #intersection of sets of interests
    l=1-(float(abs(l1-l2))/float(totlev)) #how close the level is
    c=float(len(i3))/float(totint) #how many common interests
    w1=2 #weightages
    w2=1
    match=(w1*l+w2*c)/float(w1+w2)
    return match,i3 #returning a number match between 0 to 1, highe number is more matching. i3 is set of common interests

def quest(user1,user2):
    #this function used to return the level of problem that should be given to these two users, and their interests with priorities
    l1=user1.level
    l2=user2.level
    i1=user1.interests
    i2=user2.interests
    l=int(ceil((l1+l2)/2.0)) #problem assigned more probable to have higher level
    i3=i1&i2 #intersection of sets of interests, will have more priority
    i11=i1-i3   #other intrests of user1, those which are not interests of user 2, found by difference of sets
    i22=i2-i3
    i4=set([(i,2) for i in i3]) #giving tuples of (interest,prioirty) and these have higher priority of 2 as common interests
    i5=i4 | set([(i,1) for i in i11]) | set([(i,1) for i in i22]) #taking union, the other 2 sets have tuples of lower priority as they are not in intersection of interests
    return l,i5 #returning the level of the required question, and a set of tuples of interests with priorities

def modify(user1,user2,lev,sta,totlev):
    #this function is called when users finished a problem. they either solved it (sta=1) , skipped it (sta=0), coudn't solve(sta=-1)
    #used to update points and level of users
    #sta is 1 if win/solved, 0 if tie/skip , -1 if loss/coudn't solve 
    #lev is level of problem attempted
    l1=user1.level
    l2=user2.level
    j1=(abs(lev-l1))-(lev-l1) #j1 used to subtract points on loss. its 0 if level of problem is higher than players level
    j2=(abs(lev-l2))-(lev-l2)
    dicw={1:4,2:7,3:10,4:14,5:16,6:20,7:40,8:60,9:100,10:160} #points given on solving a problem of a particular level
    if sta==1: #win ie solved problem
        user1.points+=dicw[lev] #adding the points for solving problem of level = lev
        user2.points+=dicw[lev]
        user1.update() #after incrementing points, updating level and values in database
        user2.update()
    elif sta==-1: #loss ie coudn't solve problem
        user1.points-=j1
        user2.points-=j2
        user1.update()
        user2.update()
