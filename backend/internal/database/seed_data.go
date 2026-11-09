package database

import "adaptive-assessment/internal/models"

type seedQuestion struct {
	topic, category string
	difficulty      models.Difficulty
	text            string
	opts            []string
	correct         int // index into opts
	explanation     string
}

var seedQuestionData = []seedQuestion{
	// ---------- EASY ----------
	{"Data Structures", "DSA", models.Easy, "Which data structure follows the FIFO (First In, First Out) principle?",
		[]string{"Stack", "Queue", "Tree", "Graph"}, 1, "A queue removes elements in the same order they were added."},
	{"Data Structures", "DSA", models.Easy, "Which data structure follows the LIFO (Last In, First Out) principle?",
		[]string{"Stack", "Queue", "Heap", "Array"}, 0, "A stack removes the most recently added element first."},
	{"Java", "Java", models.Easy, "Which keyword is used to inherit a class in Java?",
		[]string{"implements", "inherits", "extends", "super"}, 2, "'extends' creates a subclass of another class."},
	{"OOP", "OOP", models.Easy, "What is encapsulation in OOP?",
		[]string{"Hiding implementation using inheritance", "Bundling data and methods and restricting direct access to the data", "Creating multiple methods with the same name", "Converting one type into another"}, 1,
		"Encapsulation wraps data and behaviour in a class and controls access to it."},
	{"DBMS", "DBMS", models.Easy, "What is a primary key?",
		[]string{"A column that uniquely identifies each row", "A column that can hold duplicate values", "A key used only for sorting", "A column that stores passwords"}, 0,
		"A primary key is unique and not null, so it identifies exactly one row."},
	{"SQL", "SQL", models.Easy, "Which SQL command removes specific rows from a table?",
		[]string{"SELECT", "INSERT", "UPDATE", "DELETE"}, 3, "DELETE removes rows that match the WHERE condition."},
	{"Operating Systems", "OS", models.Easy, "What is the core component of an operating system that manages hardware and processes?",
		[]string{"Shell", "Kernel", "Compiler", "Bootloader"}, 1, "The kernel has full control over hardware, memory and processes."},
	{"Computer Networks", "Computer Networks", models.Easy, "Which protocol is primarily used to transfer web pages?",
		[]string{"FTP", "SMTP", "HTTP", "SSH"}, 2, "HTTP (HyperText Transfer Protocol) carries web page requests and responses."},
	{"System Design", "System Design", models.Easy, "What does REST stand for?",
		[]string{"Remote Execution of Server Tasks", "Representational State Transfer", "Reliable Event Streaming Technology", "Resource Encoding and Storage Tree"}, 1,
		"REST is an architectural style for designing networked APIs."},
	{"Programming", "Programming", models.Easy, "What is a function that calls itself called?",
		[]string{"Iteration", "Overloading", "Recursion", "Encapsulation"}, 2, "Recursion solves a problem by calling the same function on smaller inputs."},

	// ---------- MEDIUM ----------
	{"Data Structures", "DSA", models.Medium, "What is the time complexity of binary search on a sorted array?",
		[]string{"O(n)", "O(n log n)", "O(1)", "O(log n)"}, 3, "Each step halves the search space, giving O(log n)."},
	{"Data Structures", "DSA", models.Medium, "Which traversal of a binary search tree visits the nodes in sorted order?",
		[]string{"Preorder", "Inorder", "Postorder", "Level order"}, 1, "Inorder visits left, node, right, which yields ascending order in a BST."},
	{"Java", "Java", models.Medium, "In Java, what does the == operator compare for two String objects?",
		[]string{"Their character contents", "Their lengths", "Their memory references", "Their hash codes"}, 2, "== compares references; use equals() to compare contents."},
	{"Java", "Java", models.Medium, "Which Java collection does NOT allow duplicate elements?",
		[]string{"ArrayList", "LinkedList", "HashSet", "Vector"}, 2, "A Set stores unique elements only."},
	{"OOP", "OOP", models.Medium, "What is method overriding?",
		[]string{"Defining a method with the same name but different parameters in the same class", "A subclass providing its own implementation of a parent class method", "Hiding a method using the private keyword", "Calling a method from a static block"}, 1,
		"Overriding is runtime polymorphism: the subclass method replaces the parent's behaviour."},
	{"DBMS", "DBMS", models.Medium, "What is the main purpose of normalization?",
		[]string{"Increase redundancy to speed up reads", "Reduce data redundancy and improve data integrity", "Encrypt data at rest", "Create backups automatically"}, 1,
		"Normalization organizes tables to remove duplicate data and update anomalies."},
	{"SQL", "SQL", models.Medium, "Which SQL clause filters groups created by GROUP BY?",
		[]string{"WHERE", "ORDER BY", "HAVING", "LIMIT"}, 2, "WHERE filters rows before grouping; HAVING filters groups after aggregation."},
	{"Operating Systems", "OS", models.Medium, "What is a deadlock?",
		[]string{"A process that runs forever without blocking", "A situation where processes wait on each other forever for resources", "A crash caused by running out of memory", "A scheduling policy that favours short jobs"}, 1,
		"In a deadlock each process holds a resource and waits for one held by another."},
	{"Computer Networks", "Computer Networks", models.Medium, "Which transport protocol is connectionless and does not guarantee delivery?",
		[]string{"TCP", "UDP", "HTTP", "TLS"}, 1, "UDP sends datagrams without a handshake or delivery guarantees."},
	{"System Design", "System Design", models.Medium, "What is the main purpose of a load balancer?",
		[]string{"Store session data permanently", "Distribute incoming requests across multiple servers", "Compile application code", "Encrypt database tables"}, 1,
		"Load balancers spread traffic to improve availability and throughput."},

	// ---------- HARD ----------
	{"Data Structures", "DSA", models.Hard, "What is the worst-case time complexity of quicksort?",
		[]string{"O(n log n)", "O(n)", "O(n²)", "O(log n)"}, 2, "Consistently bad pivots (e.g., sorted input with a naive pivot) cause O(n²)."},
	{"Data Structures", "DSA", models.Hard, "Which algorithm finds shortest paths from one source in a graph with non-negative edge weights?",
		[]string{"Dijkstra's algorithm", "Kruskal's algorithm", "Depth-first search", "Floyd's cycle detection"}, 0, "Dijkstra greedily expands the closest unvisited vertex."},
	{"Java", "Java", models.Hard, "What does the volatile keyword guarantee in Java?",
		[]string{"Atomicity of compound operations", "Visibility of writes across threads", "Mutual exclusion on a block", "That the variable is stored on the stack"}, 1,
		"volatile guarantees visibility and ordering, but not atomicity of operations like i++."},
	{"OOP", "OOP", models.Hard, "Which SOLID principle says software entities should be open for extension but closed for modification?",
		[]string{"Single Responsibility", "Liskov Substitution", "Open/Closed", "Interface Segregation"}, 2, "Add new behaviour through new code, not by editing existing, tested code."},
	{"DBMS", "DBMS", models.Hard, "Which transaction isolation level prevents dirty reads but still allows non-repeatable reads?",
		[]string{"READ UNCOMMITTED", "READ COMMITTED", "REPEATABLE READ", "SERIALIZABLE"}, 1, "READ COMMITTED only reads committed data, but re-reading a row can return new values."},
	{"DBMS", "DBMS", models.Hard, "In ACID, what does the 'I' guarantee?",
		[]string{"Integrity of foreign keys", "Indexing of all columns", "Isolation between concurrent transactions", "Immediate disk writes"}, 2, "Isolation makes concurrent transactions behave as if run one after another."},
	{"SQL", "SQL", models.Hard, "What is a covering index?",
		[]string{"An index on every column of a table", "An index that contains all columns a query needs, so the table is not read", "An index that is automatically encrypted", "A temporary index created per session"}, 1,
		"If the index has every column the query uses, the database answers from the index alone."},
	{"Operating Systems", "OS", models.Hard, "Which of the following is NOT a necessary condition for deadlock?",
		[]string{"Mutual exclusion", "Hold and wait", "Circular wait", "Preemption"}, 3, "The condition is NO preemption; if resources can be preempted, deadlock can be broken."},
	{"Computer Networks", "Computer Networks", models.Hard, "What is the purpose of the TCP three-way handshake?",
		[]string{"To encrypt the connection", "To establish a connection and synchronize sequence numbers", "To resolve the server's IP address", "To measure network bandwidth"}, 1, "SYN, SYN-ACK, ACK sets up the connection and agrees on initial sequence numbers."},
	{"System Design", "System Design", models.Hard, "According to the CAP theorem, during a network partition a system must choose between which two properties?",
		[]string{"Consistency and Availability", "Consistency and Performance", "Availability and Scalability", "Partitioning and Durability"}, 0, "Under partition you can keep either consistency or availability, not both."},
	{"System Design", "System Design", models.Hard, "What does consistent hashing minimize when servers are added or removed?",
		[]string{"Network latency", "The number of keys that must be remapped", "Memory used per server", "The number of database indexes"}, 1, "Only a small fraction of keys move, which is why it's used in distributed caches."},
}