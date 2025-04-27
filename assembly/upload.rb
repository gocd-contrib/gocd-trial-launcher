#!/usr/bin/env ruby -w

require "fileutils"
require "uri"
require "json"
require "optparse"

ROOT_DIR = File.absolute_path(File.join(File.dirname(__FILE__), ".."))
PKG_RE   = /^gocd-(\d+\.\d+\.\d+)-(\d+)-(\d+|localbuild)-(osx|osx-aarch64|linux|windows)\.zip$/

def main(args=ARGV)
  opts = Opts.new

  opts.parse! args

  die opts.usage if ARGV.size != 1

  installers_dir = ARGV.first

  die "#{installers_dir.inspect} must be a directory" unless File.directory?(installers_dir)

  rm_rf(File.join(ROOT_DIR, "meta"))
  mkdir_p(File.join(ROOT_DIR, "meta"))

  rel_info = create_release_metadata(installers_dir)

  die "No installers found!" unless rel_info.size > 0

  if opts.val(:promote)
    write_to_file(File.join(ROOT_DIR, "meta", "stable.json"), rel_info.to_json)
    s3_sync ".", "test-drive", working_dir: File.join(ROOT_DIR, "meta"), cache_ctl: 60
    s3_rm "test-drive/installers", exclude: "#{info[:version]}/*"
    return
  end

  write_to_file(File.join(ROOT_DIR, "meta", "latest.json"), rel_info.to_json)

  s3_sync ".", "test-drive", working_dir: File.join(ROOT_DIR, "meta"), cache_ctl: 60

  info = rel_info[rel_info.keys.first]
  s3_sync ".", "test-drive/installers/#{info[:version]}/#{info[:build]}/", working_dir: File.join(ROOT_DIR, "installers")
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

def s3_sync(src, dest, opts={working_dir: Dir.pwd, cache_ctl: 31536000})
  dest = URI.join("s3://#{getenv!("GOCD_UPLOAD_S3_BUCKET")}", dest)
  cmd = "aws s3 sync --no-progress --acl public-read --cache-control 'max-age=#{opts[:cache_ctl]}' #{src} #{dest}"

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

def s3_rm(dest, opts={exclude: '*'})
  dest = URI.join("s3://#{getenv!("GOCD_UPLOAD_S3_BUCKET")}", dest)
  cmd = "aws s3 rm #{dest} #{dry_run? ? "--dryrun": ""} --recursive --exclude '#{opts[:exclude]}'"

  die "Failed to clean-up from #{dest} (other than '#{opts[:exclude]}')" unless system(cmd)
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
  dry_run? ? ENV.fetch(name, "dry-run-placeholder-value-for-#{hostname_safe_name(name)}") : ENV.fetch(name)
rescue
  die "This script requires the environment variable #{name.inspect}"
end

def hostname_safe_name(str)
  str.gsub(/\W/, "-").gsub("_", "-")
end

class Opts
  def initialize
    @options = {}
    @parser = OptionParser.new do |opts|
      opts.banner = "Usage: #{File.basename(__FILE__)} [options] installers_dir"
      opts.separator ""
      opts.separator "Takes exactly 1 argument:"
      opts.separator "    installers_dir        # Directory containing installers"
      opts.separator "Required Environment Variables:"
      opts.separator "    GOCD_UPLOAD_S3_BUCKET # S3 bucket to upload installers; NOT needed during dry-runs"
      opts.separator "Environment Toggles:"
      opts.separator "    DRY_RUN=y             # Don't actually upload, but ok to touch filesystem (e.g., write metadata)"
      opts.separator "    DRY_RUN=super         # Don't upload or touch the filesystem at all"
      opts.separator ""
      opts.separator "Options:"

      opts.on("-r", "--release", "Promote installers to stable") do
        @options[:promote] = true
      end

      opts.on_tail("-h", "--help", "Show this message") do
        puts self.usage
        exit 0
      end
    end
  end

  def vals
    @options.inspect
  end

  def val(key)
    @options[key.to_sym]
  end

  def parse!(args)
    @parser.parse! args
  rescue OptionParser::InvalidOption => e
    STDERR.puts "#{e.message}\n\n"
    die self.usage
  end

  def usage
    @parser.to_s
  end
end

main
