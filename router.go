package router

/*
Design

* simple
* micro wrapper on net/http
* choosable HTTP method. GET, POST...
* parsable URL parameter. "/user/:id"
	* mapping method args. "/user/:user_id/asset/:asset_id" => func(..., userId int, assetId int)

*/
