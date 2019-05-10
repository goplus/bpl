//
// https://docs.mongodb.com/manual/reference/mongodb-wire-protocol/

document = bson

MsgHeader = {/C
    int32   messageLength; // total message size, including this
    int32   requestID;     // identifier for this message
    int32   responseTo;    // requestID from the original request (used in responses from db)
    int32   opCode;        // request type - see table below
}

OP_UPDATE = {/C
	int32     ZERO;               // 0 - reserved for future use
	cstring   fullCollectionName; // "dbname.collectionname"
	int32     flags;              // bit vector. see below
	document  selector;           // the query to select the document
	document  update;             // specification of the update to perform
}

OP_INSERT = {/C
	int32      flags;              // bit vector - see below
	cstring    fullCollectionName; // "dbname.collectionname"
	document*  documents;          // one or more documents to insert into the collection
}

OP_QUERY = {/C
	int32     flags;                  // bit vector of query options.  See below for details.
	cstring   fullCollectionName;     // "dbname.collectionname"
	int32     numberToSkip;           // number of documents to skip
	int32     numberToReturn;         // number of documents to return
		                              //  in the first OP_REPLY batch
	document  query;                  // query object.  See below for details.
	document? returnFieldsSelector;   // Optional. Selector indicating the fields
		                              //  to return.  See below for details.
}

OP_GET_MORE = {/C
	int32     ZERO;               // 0 - reserved for future use
	cstring   fullCollectionName; // "dbname.collectionname"
	int32     numberToReturn;     // number of documents to return
	int64     cursorID;           // cursorID from the OP_REPLY
}

OP_DELETE = {/C
	int32     ZERO;               // 0 - reserved for future use
	cstring   fullCollectionName; // "dbname.collectionname"
	int32     flags;              // bit vector - see below for details.
	document  selector;           // query object.  See below for details.
}

OP_KILL_CURSORS = {/C
	int32     ZERO;              // 0 - reserved for future use
	int32     numberOfCursorIDs; // number of cursorIDs in message
	int64*    cursorIDs;         // sequence of cursorIDs to close
}

OP_MSG = {/C
	cstring   message; // message for the database
}

OP_REPLY = {/C
	int32     responseFlags;  // bit vector - see details below
	int64     cursorID;       // cursor id if client needs to do get more's
	int32     startingFrom;   // where in the cursor this reply is starting
	int32     numberReturned; // number of documents in the reply
	document* documents;      // documents
}

OP_REQ = {/C
	cstring  dbName;
	cstring  cmd;
	document param;
}

OP_RET = {/C
	document ret;
}

Message = {
	header MsgHeader   // standard message header
	_body  [header.messageLength - sizeof(MsgHeader)]byte
	eval _body do case header.opCode {
		1:    OP_REPLY    // Reply to a client request. responseTo is set.
		1000: OP_MSG      // Generic msg command followed by a string.
		2001: OP_UPDATE
		2002: OP_INSERT
		2004: OP_QUERY
		2005: OP_GET_MORE // Get more data from a query. See Cursors.
		2006: OP_DELETE
		2007: OP_KILL_CURSORS // Notify database that the client has finished with the cursor.
		2010: OP_REQ
		2011: OP_RET
		default: let body = _body
	}
}

doc = *(Message dump)
