use axum::{extract, Router, routing::{post}};
use tokio;
use std::fs;
use serde::{Deserialize, Serialize};
// simple database for notes: name string, text string. stored in json files

#[derive(Deserialize, Serialize)]
struct Note{
    Name: String,
    Text: String,
}

async fn store_note(extract::Json(payload): extract::Json<Note>){
    let note = serde_json::to_vec(&payload).unwrap();
    println!("{:?}", note);
    fs::write("output.json", note).unwrap();    
} 

#[tokio::main]
async fn main() {
    let app = Router::new().route("/save", post(store_note));

    let listener = tokio::net::TcpListener::bind("localhost:8091").await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
