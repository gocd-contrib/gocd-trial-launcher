#!/usr/bin/env ruby -w

require "fileutils"
require "uri"
require "json"

ROOT_DIR = File.absolute_path(File.join(File.dirname(__FILE__), ".."))
PKG_RE   = /^gocd-(\d+\.\d+\.\d+)-(\d+)-(\d+|localbuild)-(osx|linux|windows)\.zip$/

def main(args=ARGV)
  die usage if args.size != 1

  installers_dir = args.first

  die usage unless File.directory?(installers_dir)

  rm_rf(File.join(ROOT_DIR, "meta"))
  mkdir_p(File.join(ROOT_DIR, "meta"))

  rel_info = create_release_metadata(installers_dir)

  die "No installers found!" unless rel_info.size > 0

  write_to_file(File.join(ROOT_DIR, "meta", "latest.json"), rel_info.to_json)

  s3_sync ".", "/", working_dir: File.join(ROOT_DIR, "meta")

  info = rel_info[rel_info.keys.first]
  s3_sync ".", "installers/#{info[:version]}/#{info[:build]}/", working_dir: File.join(ROOT_DIR, "installers")
end

def create_release_metadata(src_dir)
  Dir.glob(File.join(src_dir, "gocd-*.zip")).inject({}) { |memo, file|
    name = File.basename(file)

    if m = PKG_RE.match(name)
      platform = m[4]
      memo[platform] = { version: m[1], build: m[2], trialbuild: m[3], name: name, size: file_sz_in_mb(file) }
    end

    memo
  }
end

def usage
  %Q{USAGE:
    File.basename(__FILE__) installers_dir

    Takes exactly 1 argument:
      installers_dir        # Directory containing installers

    Required Environment Variables:
      GOCD_UPLOAD_S3_BUCKET # S3 bucket to upload installers; NOT needed during dry-runs

    Environment Toggles:
      DRY_RUN=y             # Don't actually upload, but ok to touch filesystem (e.g., write metadata)
      DRY_RUN=super         # Don't upload or touch the filesystem at all
  }.strip
end

def write_to_file(path, content)
  if super_dry?
    dry_log "Would write to file #{path.inspect} content: #{content.inspect}"
  else
    File.open(path, "w") { |file|
      file.puts content
    }
  end
end

def rm_rf(dir)
  if super_dry?
    dry_log "Would remove directory #{dir.inspect}"
  else
    FileUtils.rm_rf dir
  end
end

def mkdir_p(dir)
  if super_dry?
    dry_log "Would create directory #{dir.inspect}"
  else
    FileUtils.mkdir_p dir
  end
end

def s3_sync(src, dest, opts={working_dir: Dir.pwd})
  dest = URI.join("s3://#{getenv!("GOCD_UPLOAD_S3_BUCKET")}", dest)
  cmd = "aws s3 sync --no-progress --acl public-read --cache-control 'max-age=31536000' #{src} #{dest}"

  if dry_run?
    dry_log "From dir: #{opts[:working_dir]}"
    dry_log "Would exec: #{cmd.inspect}"
    return true
  else
    Dir.chdir(opts[:working_dir]) {
      die "Failed to upload #{File.absolute_path(src)} to #{dest}" unless system(cmd)
    }
  end
end

def die(msg)
  STDERR.puts msg
  exit 1
end

def dry_log(msg)
  STDOUT.puts("[DRY RUN] #{msg}")
end

def file_sz_in_mb(file)
  "%.1f MB" % (File.size(file).to_f / 2**20)
end

def super_dry?
  dry_run? && "super" == ENV["DRY_RUN"]
end

def dry_run?
  ENV.has_key?("DRY_RUN")
end

def getenv!(name)
  dry_run? ? ENV.fetch(name, "dry-run-value-for-#{name.gsub("_", "-")}") : ENV.fetch(name)
rescue
  die "This script require the environment variable #{name.inspect}"
end

main
