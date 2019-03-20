#!/usr/bin/env ruby -w

require "fileutils"
require "uri"
require "json"

ROOT_DIR = File.absolute_path(File.join(File.dirname(__FILE__), ".."))
PKG_RE   = /^gocd-(\d+\.\d+\.\d+)-(\d+)-(\d+|localbuild)-(osx|linux|windows)\.zip$/

def main(args=ARGV)
  die "Requires exactly 1 argument; path to installers directory" if args.size != 1

  installers_dir = args.first

  die "installers path #{installers_dir.inspect} must be a directory" unless File.directory?(installers_dir)

  FileUtils.rm_rf(File.join(ROOT_DIR, "meta"))
  FileUtils.mkdir_p(File.join(ROOT_DIR, "meta"))

  rel_info = create_release_metadata(installers_dir)

  die "No installers found!" unless rel_info.size > 0

  File.open(File.join(ROOT_DIR, "meta", "latest.json"), "w") { |file|
    file.puts rel_info.to_json
  }

  Dir.chdir(File.join(ROOT_DIR, "meta")) {
    s3_sync "latest.json", "/"
  }

  Dir.chdir(File.join(ROOT_DIR, "installers")) {
    info = rel_info[rel_info.keys.first]

    s3_sync ".", "installers/#{info[:version]}/#{info[:build]}/"
  }
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

def s3_sync(src, dest)
  dest = URI.join("s3://#{getenv!("GOCD_UPLOAD_S3_BUCKET")}", dest)
  cmd = "aws s3 sync --no-progress --acl public-read --cache-control 'max-age=31536000' #{src} #{dest}"

  if dry_run?
    puts "[DRY RUN] from dir: #{Dir.pwd}"
    puts "[DRY RUN] exec: #{cmd.inspect}"
    return true
  else
    die "Failed to upload #{File.absolute_path(src)} to #{dest}" unless system(cmd)
  end
end

def die(msg)
  STDERR.puts msg
  exit 1
end

def file_sz_in_mb(file)
  "%.1f MB" % (File.size(file).to_f / 2**20)
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
