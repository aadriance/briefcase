use clap::{Parser, Subcommand};
use regex::Regex;
use std::env;
use std::fs;
use std::io::{self, Read, Write};
use std::path::PathBuf;

const VERSION: &str = "0.5.0";

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
    /// Show information about the temp directory used by briefcase
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

struct TempDir {
    path: PathBuf,
    env_var: String,
}

#[derive(Debug)]
enum BriefcaseError {
    Io(io::Error),
    InvalidEntry(String),
    EntryNotFound(String),
}

impl From<io::Error> for BriefcaseError {
    fn from(error: io::Error) -> Self {
        BriefcaseError::Io(error)
    }
}

type Result<T> = std::result::Result<T, BriefcaseError>;

fn main() -> Result<()> {
    let cli = Cli::parse();

    match &cli.command {
        Commands::Version => version(),
        Commands::Info => info()?,
        Commands::Set { name, value } => set(name, value)?,
        Commands::Get { name } => get(name)?,
        Commands::Purge { force } => purge(*force)?,
        Commands::Remove { name } => remove(name)?,
        Commands::List => list()?,
    }

    Ok(())
}

// Utility functions

fn get_temp_dir() -> TempDir {
    let env_vars = ["BRIEFCASE_DIR", "TEMP", "TMPDIR"];

    env_vars
        .iter()
        .find_map(|&env_var| {
            env::var(env_var).ok().map(|dir| TempDir {
                path: PathBuf::from(dir),
                env_var: env_var.to_string(),
            })
        })
        .unwrap_or_else(|| TempDir {
            path: PathBuf::from("/tmp"),
            env_var: "N/A".to_string(),
        })
}

fn get_briefcase_dir_name() -> String {
    env::var("BRIEFCASE_DIRNAME").unwrap_or_else(|_| "briefcase".to_string())
}

fn get_briefcase_dir() -> PathBuf {
    get_temp_dir().path.join(get_briefcase_dir_name())
}

fn is_valid_entry(entry: &str) -> bool {
    let re = Regex::new(r"^[a-zA-Z][a-zA-Z0-9_]*$").unwrap();
    re.is_match(entry)
}

// Command functions

fn version() {
    println!("Briefcase {}", VERSION);
}

fn info() -> Result<()> {
    let temp_info = get_temp_dir();
    let dir_name = get_briefcase_dir_name();
    println!("\tTemp Dir: {}", temp_info.path.display());
    println!("\tSourced From: {}", temp_info.env_var);
    println!("\tBriefcase Directory Name: {}", dir_name);
    Ok(())
}

fn set(name: &str, value: &str) -> Result<()> {
    if !is_valid_entry(name) {
        return Err(BriefcaseError::InvalidEntry(name.to_string()));
    }

    let briefcase = get_briefcase_dir();
    fs::create_dir_all(&briefcase)?;

    let file_path = briefcase.join(name);
    fs::write(file_path, value)?;
    Ok(())
}

fn get(name: &str) -> Result<()> {
    if !is_valid_entry(name) {
        return Err(BriefcaseError::InvalidEntry(name.to_string()));
    }

    let file_path = get_briefcase_dir().join(name);
    let mut file = fs::File::open(&file_path).map_err(|e| {
        if e.kind() == io::ErrorKind::NotFound {
            BriefcaseError::EntryNotFound(name.to_string())
        } else {
            BriefcaseError::Io(e)
        }
    })?;

    let mut contents = String::new();
    file.read_to_string(&mut contents)?;
    print!("{}", contents);
    Ok(())
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

    let briefcase = get_briefcase_dir();
    fs::remove_dir_all(briefcase)?;
    println!("Briefcase data purged successfully");
    Ok(())
}

fn remove(name: &str) -> Result<()> {
    if !is_valid_entry(name) {
        return Err(BriefcaseError::InvalidEntry(name.to_string()));
    }

    let file_path = get_briefcase_dir().join(name);
    fs::remove_file(file_path)?;
    println!("Removed {}", name);
    Ok(())
}

fn list() -> Result<()> {
    let briefcase = get_briefcase_dir();
    let entries = fs::read_dir(briefcase)?;

    for entry in entries {
        let entry = entry?;
        println!("{}", entry.file_name().to_string_lossy());
    }
    Ok(())
}

// Error handling

impl std::fmt::Display for BriefcaseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            BriefcaseError::Io(err) => write!(f, "IO error: {}", err),
            BriefcaseError::InvalidEntry(entry) => write!(f, "Invalid entry name: {}", entry),
            BriefcaseError::EntryNotFound(entry) => write!(f, "Entry not found: {}", entry),
        }
    }
}

impl std::error::Error for BriefcaseError {}
