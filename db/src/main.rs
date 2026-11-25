use axum::{Json, Router, extract::{self, State, Path}, routing::post, routing::get, routing::delete};
use tokio;
use uuid::Uuid;
use std::{fs::{self, create_dir_all}, path::PathBuf};
use std::env;
use serde::{Deserialize, Serialize};

#[derive(Deserialize, Serialize)]
struct Note{
    Name: String,
    Text: String,
    Id: Uuid,
}

#[derive(Clone)]
struct Config {
    pub host: String,
    pub db_port: u16,
    pub db_dir: PathBuf,
}

impl Config {
    pub fn from_env() -> Result<Self, Box<dyn std::error::Error>> {
        dotenvy::dotenv().ok();

        let host = env::var("HOST")
            .expect("Переменная окружения HOST должна быть установлена");

        let db_dir_str = env::var("DB_DIR")
            .expect("Переменная окружения DB_DIR должна быть установлена");
        let db_port = env::var("DB_PORT")
            .expect("Переменная окружения DB_PORT должна быть установлена")
            .parse::<u16>()
            .expect("DB_PORT должен быть числом от 0 до 65535");
        let db_dir = PathBuf::from(db_dir_str);
        
        Ok(Self {host, db_port, db_dir})
    }
}

async fn store_note(State(config): State<Config>, extract::Json(mut payload): extract::Json<Note>){
    let uuid = Uuid::new_v4();
    payload.Id = uuid;
    let note = serde_json::to_vec(&payload).unwrap();
    let filename = format!("{}.json", uuid); 
    let note_path = config.db_dir.join(filename);
    fs::write(note_path, note).unwrap();    
} 

async fn list_notes(State(config): State<Config>) -> Json<Vec<Note>>{
    let mut paths = tokio::fs::read_dir(config.db_dir).await.unwrap();
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

async fn delete_note(State(config): State<Config>, Path(id): Path<Uuid>) {
    let filename = format!("{}.json", id);
    let file = config.db_dir.join(filename);
    tokio::fs::remove_file(file).await.unwrap();
}

fn app(config: Config) -> Router {
    Router::new()
        .route("/save", post(store_note))
        .route("/notes", get(list_notes))
        .route("/delete/{id}", delete(delete_note)).with_state(config)
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>>{
    let config = Config::from_env()?;
    
    let data_dir = std::env::current_dir()?.join(&config.db_dir);
    create_dir_all(&data_dir)?; 

    let addr = format!("{}:{}", config.host, config.db_port);
    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app(config)).await?;
    
    Ok(())
}
