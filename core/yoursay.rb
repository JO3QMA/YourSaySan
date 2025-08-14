# frozen_string_literal: true

require 'discordrb'
require 'config'
require 'yaml'
require 'erb'

# YourSaySanモジュール
module YourSaySan
  # Load config.yml with ERB + YAML so ENV values can be embedded
  CONFIG = begin
    erb = ERB.new(File.read('config.yml'))
    yaml = YAML.safe_load(erb.result, aliases: true)
    Config.load_and_set_settings(yaml)
  end
  BOT = Discordrb::Commands::CommandBot.new(
    token: CONFIG.bot.token,
    client_id: CONFIG.bot.client_id,
    prefix: CONFIG.bot.prefix,
    ignore_bots: true
  )

  # Module define

  module Commands; end
  module Events; end

  # コマンドとイベントを登録する
  def self.load_modules
    puts '[Bot] モジュールの読み込み開始'

    # Bot Commands
    Dir['./core/modules/commands/*.rb'].sort.each do |file|
      puts "[Bot] Load Command: #{file}"
      require file
    end
    Commands.constants.each do |mod|
      BOT.include! Commands.const_get mod
    end

    # Bot Events
    Dir['./core/modules/events/*.rb'].sort.each do |file|
      puts "[Bot] Load Event: #{file}"
      require file
    end
    Events.constants.each do |mod|
      BOT.include! Events.const_get mod
    end
  end

  def self.run
    load_modules
    BOT.run
  rescue StandardError => e
    puts "Bot実行中にエラーが発生しました: #{e.message}"
    puts e.backtrace.join("\n")
  end
end
