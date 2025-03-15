const daysLeaders = [
    [
        { nickname: "Arjun Mahesh", total_winnings: 419682 },
        { nickname: "Pavan Srivastava", total_winnings: 366297 },
        { nickname: "Vishal Patel", total_winnings: 342168 },
        { nickname: "Deepak Singh", total_winnings: 326745 },
        { nickname: "Ravi Sharma", total_winnings: 301154 },
        { nickname: "Priyanka Sharma", total_winnings: 287417 },
        { nickname: "Isha Mali", total_winnings: 278905 },
        { nickname: "Karishma Reddy", total_winnings: 203580 },
        { nickname: "Rajesh Kumar", total_winnings: 154736 },
        { nickname: "Nisha Yadav", total_winnings: 145820 }
    ],
    [
        { nickname: "Raghav Sharma", total_winnings: 432672 },
        { nickname: "Ayesha Khan", total_winnings: 389572 },
        { nickname: "Sonya Chauhan", total_winnings: 367921 },
        { nickname: "Rahul Jain", total_winnings: 298604 },
        { nickname: "Neha Gupta", total_winnings: 290734 },
        { nickname: "Simran Kaur", total_winnings: 185703 },
        { nickname: "Mahesh Verma", total_winnings: 250387 },
        { nickname: "Nina Patel", total_winnings: 304918 },
        { nickname: "Kartik Agrawal", total_winnings: 212579 },
        { nickname: "Latifa Mahmood", total_winnings: 175014 }
    ],
    [
        { nickname: "Ria Mehta", total_winnings: 368501 },
        { nickname: "Aditi Verma", total_winnings: 315839 },
        { nickname: "Karan Singh", total_winnings: 299612 },
        { nickname: "Ananya Soni", total_winnings: 342014 },
        { nickname: "Nikhil Yadav", total_winnings: 285107 },
        { nickname: "Manisha Gupta", total_winnings: 278483 },
        { nickname: "Kabir Rajput", total_winnings: 157649 },
        { nickname: "Vikas Rao", total_winnings: 249560 },
        { nickname: "Preeti Mishra", total_winnings: 204923 },
        { nickname: "Anil Kumar", total_winnings: 212445 }
    ],
    [
        { nickname: "Sanjana Desai", total_winnings: 404289 },
        { nickname: "Vikram Sharma", total_winnings: 347583 },
        { nickname: "Aryan Sood", total_winnings: 327110 },
        { nickname: "Rahul Patil", total_winnings: 321054 },
        { nickname: "Tanuja Das", total_winnings: 265910 },
        { nickname: "Pooja Mehra", total_winnings: 213812 },
        { nickname: "Manish Kumar", total_winnings: 236548 },
        { nickname: "Harshit Agarwal", total_winnings: 145632 },
        { nickname: "Neha Jain", total_winnings: 198726 },
        { nickname: "Kavita Yadav", total_winnings: 158264 }
    ],
    [
        { nickname: "Rishi Kumar", total_winnings: 428310 },
        { nickname: "Neelam Rao", total_winnings: 175383 },
        { nickname: "Akash Saini", total_winnings: 299220 },
        { nickname: "Tushar Jain", total_winnings: 326548 },
        { nickname: "Simran Singh", total_winnings: 244913 },
        { nickname: "Snehal Ghosh", total_winnings: 218077 },
        { nickname: "Shikha Jain", total_winnings: 248442 },
        { nickname: "Shivani Yadav", total_winnings: 295619 },
        { nickname: "Abhishek Kapoor", total_winnings: 379014 },
        { nickname: "Aarti Verma", total_winnings: 185620 }
    ],
    [
        { nickname: "Kiran Sharma", total_winnings: 367743 },
        { nickname: "Harsh Rajput", total_winnings: 251175 },
        { nickname: "Rohit Kumar", total_winnings: 292364 },
        { nickname: "Simran Sharma", total_winnings: 234651 },
        { nickname: "Shweta Chaudhary", total_winnings: 375021 },
        { nickname: "Abhinav Yadav", total_winnings: 248715 },
        { nickname: "Rekha Mehta", total_winnings: 314872 },
        { nickname: "Manav Agarwal", total_winnings: 357264 },
        { nickname: "Nisha Yadav", total_winnings: 265460 },
        { nickname: "Sanjay Deshmukh", total_winnings: 211385 }
    ],
    [
        { nickname: "Rajiv Kumar", total_winnings: 431120 },
        { nickname: "Amandeep Kaur", total_winnings: 318514 },
        { nickname: "Nitin Yadav", total_winnings: 304103 },
        { nickname: "Preeti Soni", total_winnings: 262941 },
        { nickname: "Gaurav Singh", total_winnings: 235492 },
        { nickname: "Tanya Gupta", total_winnings: 305671 },
        { nickname: "Ankit Sharma", total_winnings: 250823 },
        { nickname: "Alok Mishra", total_winnings: 212376 },
        { nickname: "Sandeep Jain", total_winnings: 178210 },
        { nickname: "Sheetal Mehta", total_winnings: 210345 }
    ],
    [
        { nickname: "Priya Sharma", total_winnings: 398573 },
        { nickname: "Karan Yadav", total_winnings: 350987 },
        { nickname: "Harsha Gupta", total_winnings: 283429 },
        { nickname: "Ramesh Kumar", total_winnings: 311789 },
        { nickname: "Meena Singh", total_winnings: 267487 },
        { nickname: "Umesh Verma", total_winnings: 235684 },
        { nickname: "Shilpa Reddy", total_winnings: 256174 },
        { nickname: "Ravi Mehra", total_winnings: 194367 },
        { nickname: "Rekha Patel", total_winnings: 157612 },
        { nickname: "Vinod Soni", total_winnings: 188450 }
    ],
    [
        { nickname: "Aastha Sharma", total_winnings: 423672 },
        { nickname: "Rajat Soni", total_winnings: 299231 },
        { nickname: "Neha Agarwal", total_winnings: 310944 },
        { nickname: "Kapil Gupta", total_winnings: 276552 },
        { nickname: "Amrita Reddy", total_winnings: 263101 },
        { nickname: "Sunil Sharma", total_winnings: 312398 },
        { nickname: "Manju Verma", total_winnings: 236843 },
        { nickname: "Poonam Kaur", total_winnings: 214223 },
        { nickname: "Arvind Rajput", total_winnings: 198328 },
        { nickname: "Rinku Yadav", total_winnings: 175598 }
    ],
    [
        { nickname: "Anjali Verma", total_winnings: 420835 },
        { nickname: "Nitin Kumar", total_winnings: 358239 },
        { nickname: "Amit Mehra", total_winnings: 289856 },
        { nickname: "Sonal Yadav", total_winnings: 274823 },
        { nickname: "Hitesh Jain", total_winnings: 204940 },
        { nickname: "Kusum Kaur", total_winnings: 216876 },
        { nickname: "Sudhir Agarwal", total_winnings: 298674 },
        { nickname: "Pooja Soni", total_winnings: 227111 },
        { nickname: "Ramesh Yadav", total_winnings: 180212 },
        { nickname: "Saurabh Gupta", total_winnings: 168903 }
    ]
];

const weeksLeaders = [
    [
        { nickname: "Arjun Verma", total_winnings: 876453 },
        { nickname: "Priya Yadav", total_winnings: 842547 },
        { nickname: "Rohit Gupta", total_winnings: 798230 },
        { nickname: "Ramesh Mehta", total_winnings: 762403 },
        { nickname: "Ayesha Singh", total_winnings: 741912 },
        { nickname: "Karan Kapoor", total_winnings: 718520 },
        { nickname: "Simran Kaur", total_winnings: 692187 },
        { nickname: "Rahul Patel", total_winnings: 659470 },
        { nickname: "Manisha Arora", total_winnings: 624100 },
        { nickname: "Vikash Yadav", total_winnings: 587700 }
    ],
    [
        { nickname: "Pankaj Rajput", total_winnings: 865412 },
        { nickname: "Nisha Gupta", total_winnings: 834112 },
        { nickname: "Karan Sharma", total_winnings: 810542 },
        { nickname: "Sonal Yadav", total_winnings: 775643 },
        { nickname: "Tanvi Mehra", total_winnings: 740908 },
        { nickname: "Deepak Patel", total_winnings: 703411 },
        { nickname: "Karishma Kapoor", total_winnings: 678372 },
        { nickname: "Neha Jain", total_winnings: 647290 },
        { nickname: "Akash Kumar", total_winnings: 618481 },
        { nickname: "Priyanka Mehta", total_winnings: 593055 }
    ],
    [
        { nickname: "Manish Yadav", total_winnings: 891332 },
        { nickname: "Aarti Kapoor", total_winnings: 845923 },
        { nickname: "Gaurav Singh", total_winnings: 806120 },
        { nickname: "Shubham Reddy", total_winnings: 773410 },
        { nickname: "Raghav Jain", total_winnings: 741703 },
        { nickname: "Kavita Mehra", total_winnings: 711927 },
        { nickname: "Sandeep Arora", total_winnings: 688159 },
        { nickname: "Neelam Reddy", total_winnings: 651443 },
        { nickname: "Harshita Yadav", total_winnings: 621568 },
        { nickname: "Anil Gupta", total_winnings: 598234 }
    ],
    [
        { nickname: "Arvind Yadav", total_winnings: 863034 },
        { nickname: "Priya Singh", total_winnings: 830907 },
        { nickname: "Vinay Patel", total_winnings: 792421 },
        { nickname: "Ria Kapoor", total_winnings: 767211 },
        { nickname: "Kiran Gupta", total_winnings: 745213 },
        { nickname: "Abhishek Mehra", total_winnings: 718110 },
        { nickname: "Rekha Yadav", total_winnings: 683240 },
        { nickname: "Nikhil Verma", total_winnings: 649356 },
        { nickname: "Sanjana Kapoor", total_winnings: 618905 },
        { nickname: "Pooja Soni", total_winnings: 590430 }
    ]
];

const allTimeLeaders = [
    { nickname: "Aarav Bhatia", total_winnings: 1565307 },
    { nickname: "Neha Joshi", total_winnings: 1451212 },
    { nickname: "Raghav Chauhan", total_winnings: 1346928 },
    { nickname: "Sonal Reddy", total_winnings: 1221107 },
    { nickname: "Karan Verma", total_winnings: 1120455 },
    { nickname: "Priya Malik", total_winnings: 1037123 },
    { nickname: "Ankit Soni", total_winnings: 987654 },
    { nickname: "Simran Patel", total_winnings: 963109 },
    { nickname: "Deepak Agarwal", total_winnings: 948722 },
    { nickname: "Pooja Nair", total_winnings: 936271 }
];

const getLeadersDates = () => {

    const getDay = () => {
        return new Date().getDate(); 
    };

    const getCurrentWeekOfMonth = () => {
        const now = new Date();
        const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
        
        return Math.ceil((now.getDate() + startOfMonth.getDay()) / 7) - 1;
    };

    const getDailyLeaders = (list) => {
        const dayIndex = (getDay() - 1) % daysLeaders.length;
        const currentLeaders = daysLeaders[dayIndex]; 
        
        return list.concat(currentLeaders)
                   .sort((a, b) => b.total_winnings - a.total_winnings)
                   .slice(0, 10);
    };

    const getWeeklyLeaders = (list) => {
        const weekIndex = getCurrentWeekOfMonth() % weeksLeaders.length; 
        const currentWeek = weeksLeaders[weekIndex];

        return list.concat(currentWeek)
                   .sort((a, b) => b.total_winnings - a.total_winnings)
                   .slice(0, 10);
    };

    const getAllLeaders = (list) => {
        return list.concat(allTimeLeaders)
                   .sort((a, b) => b.total_winnings - a.total_winnings)
                   .slice(0, 10);
    };

    return { getDailyLeaders, getWeeklyLeaders, getAllLeaders };
};

export { getLeadersDates };
