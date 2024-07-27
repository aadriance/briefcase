use anyhow::{Context, Result};
use clap::{Parser, Subcommand};
use redb::{Database, ReadableTable, ReadableTableMetadata, TableDefinition};
use regex::Regex;
use std::fs;
use std::io::{self, Write};
use std::path::PathBuf;

const DB_NAME: &str = "briefcase.redb";
const TABLE: TableDefinition<&str, &str> = TableDefinition::new("briefcase");

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Show the version of briefcase
    Version,
    /// Show information about the database used by briefcase
    Info,
    /// Set a briefcase variable
    Set {
        /// Name of the variable
        name: String,
        /// Value of the variable
        value: String,
    },
    /// Get a briefcase variable
    Get {
        /// Name of the variable
        name: String,
    },
    /// Purge briefcase data
    Purge {
        /// Force purge without confirmation
        #[arg(long)]
        force: bool,
    },
    /// Remove a briefcase variable
    Remove {
        /// Name of the variable
        name: String,
    },
    /// List briefcase entries
    List,
}

fn main() -> Result<()> {
    let cli = Cli::parse();

    match &cli.command {
        Commands::Version => version(),
        Commands::Info => info(),
        Commands::Set { name, value } => set(name, value),
        Commands::Get { name } => get(name),
        Commands::Purge { force } => purge(*force),
        Commands::Remove { name } => remove(name),
        Commands::List => list(),
    }
}

fn get_db_path() -> PathBuf {
    let home_dir = dirs::home_dir().expect("Unable to determine home directory");
    home_dir.join(".briefcase").join(DB_NAME)
}

fn open_db() -> Result<Database> {
    let db_path = get_db_path();
    fs::create_dir_all(db_path.parent().unwrap()).context("Failed to create database directory")?;
    Database::create(db_path).context("Failed to open database")
}

fn is_valid_entry(entry: &str) -> bool {
    let re = Regex::new(r"^[a-zA-Z][a-zA-Z0-9_]*$").unwrap();
    re.is_match(entry)
}

// Command functions

fn version() -> Result<()> {
    println!("Briefcase {}", std::env!("CARGO_PKG_VERSION"));
    Ok(())
}

fn info() -> Result<()> {
    let db_path = get_db_path();
    println!("Database path: {}", db_path.display());
    let db = open_db()?;
    let reader = db.begin_read()?;
    let table = reader.open_table(TABLE)?;
    println!("Number of entries: {}", table.len()?);
    Ok(())
}

fn set(name: &str, value: &str) -> Result<()> {
    if !is_valid_entry(name) {
        anyhow::bail!("Invalid entry name: {}", name);
    }

    let db = open_db()?;
    let write_txn = db.begin_write()?;
    {
        let mut table = write_txn.open_table(TABLE)?;
        table.insert(name, value)?;
    }
    write_txn.commit()?;
    println!("Set {} = {}", name, value);
    Ok(())
}

fn get(name: &str) -> Result<()> {
    if !is_valid_entry(name) {
        anyhow::bail!("Invalid entry name: {}", name);
    }

    let db = open_db()?;
    let reader = db.begin_read()?;
    let table = reader.open_table(TABLE)?;

    match table.get(name)? {
        Some(value) => {
            print!("{}", value.value());
            Ok(())
        }
        None => anyhow::bail!("Entry not found: {}", name),
    }
}

fn purge(force: bool) -> Result<()> {
    if !force {
        print!("Are you sure you want to delete all briefcase data? (y/n) ");
        io::stdout().flush()?;
        let mut input = String::new();
        io::stdin().read_line(&mut input)?;
        if input.trim().to_lowercase() != "y" {
            println!("Exiting without deleting data");
            return Ok(());
        }
    }

    let db_path = get_db_path();
    fs::remove_file(db_path).context("Failed to remove database file")?;
    println!("Briefcase data purged successfully");
    Ok(())
}

fn remove(name: &str) -> Result<()> {
    if !is_valid_entry(name) {
        anyhow::bail!("Invalid entry name: {}", name);
    }

    let db = open_db()?;
    let write_txn = db.begin_write()?;
    {
        let mut table = write_txn.open_table(TABLE)?;
        table.remove(name)?;
    }
    write_txn.commit()?;
    Ok(())
}

fn list() -> Result<()> {
    let db = open_db()?;
    let reader = db.begin_read()?;
    let table = reader.open_table(TABLE)?;

    for result in table.iter()? {
        let (key, value) = result?;
        println!("{} = {}", key.value(), value.value());
    }
    Ok(())
}
