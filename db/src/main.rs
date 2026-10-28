use axum::{Json, Router, extract::{self, State}, routing::post, routing::get};
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

async fn list_notes(State(data_dir): State<PathBuf>) -> Json<Vec<Note>>{
    let mut paths = tokio::fs::read_dir(data_dir).await.unwrap();
    let mut data: Vec<Note> = Vec::new();

    loop {
        if let Some(entry) = paths.next_entry().await.unwrap() {
            let path = entry.path();
            if path.is_file() {
                let file = tokio::fs::read(path).await.unwrap();
                data.push(serde_json::from_slice(&file).unwrap());
            }
        } else{
            break;
        }

    }
    Json(data)
}

fn app() -> Router {
    dotenvy::dotenv().ok();
    let dbdir = env::var("DB_DIR").unwrap();
    let mut data_dir = env::current_dir().unwrap();
    data_dir.push(dbdir);
    create_dir_all(&data_dir).unwrap();

    Router::new()
        .route("/save", post(store_note)).with_state(data_dir.clone())
        .route("/notes", get(list_notes)).with_state(data_dir)
}

#[tokio::main]
async fn main() {
    let listener = tokio::net::TcpListener::bind("localhost:8091").await.unwrap();
    axum::serve(listener, app()).await.unwrap();
}
