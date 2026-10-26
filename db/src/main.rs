use axum::{Router, extract::{self, State}, routing::post};
use tokio;
use uuid::Uuid;
use std::{fs::{self, create_dir_all}, path::PathBuf};
use std::env;
use serde::{Deserialize, Serialize};

#[derive(Deserialize, Serialize)]
struct Note{
    Name: String,
    Text: String,
}

async fn store_note(State(data_dir): State<PathBuf>, extract::Json(payload): extract::Json<Note>){
    let note = serde_json::to_vec(&payload).unwrap();
    let filename = format!("{}.json", Uuid::new_v4()); 
    let note_path = data_dir.join(filename);
    fs::write(note_path, note).unwrap();    
} 

#[tokio::main]
async fn main() {
    dotenvy::dotenv().ok();

    let dbdir = env::var("DB_DIR").unwrap();
    let mut data_dir = env::current_dir().unwrap();
    data_dir.push(dbdir);
    create_dir_all(&data_dir).unwrap();

    let app = Router::new().route("/save", post(store_note)).with_state(data_dir);
    let listener = tokio::net::TcpListener::bind("localhost:8091").await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
