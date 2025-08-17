# frozen_string_literal: true

require 'discordrb'
require 'config'
require 'yaml'
require 'erb'
require 'logger'
require_relative 'voicevox'
require_relative 'speaker_manager'

# YourSaySanモジュール
module YourSaySan
  # Load config.yml with ERB + YAML so ENV values can be embedded
  CONFIG = begin
    erb = ERB.new(File.read('config.yml'))
    yaml = YAML.safe_load(erb.result, aliases: true)
    Config.load_and_set_settings(yaml)
  end
  BOT = Discordrb::Bot.new(
    token: CONFIG.bot.token,
    client_id: CONFIG.bot.client_id,
    ignore_bots: true
  )

  # Module define

  module Commands; end
  module Events; end

  # Shared states
  @text_channels = []
  @voicevox = VoiceVox.new(CONFIG, Logger.new($stdout)) rescue nil
  @speaker_manager = SpeakerManager.new(CONFIG, Logger.new($stdout), @voicevox) rescue nil

  def self.text_channels
    @text_channels
  end

  def self.voicevox
    @voicevox
  end

  def self.speaker_manager
    @speaker_manager
  end

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
      mod_ref = Events.const_get mod
      BOT.include! mod_ref
      # イベントモジュールが setup を提供している場合のみ呼ぶ（テストで差し替え可能）
      mod_ref.setup(BOT, @text_channels, @voicevox, CONFIG) if mod_ref.respond_to?(:setup)
    end

    # 各コマンドモジュールのコマンド登録を実行
    register_slash_commands_from_modules
  end

  def self.register_slash_commands_from_modules
    puts '[Bot] コマンドの登録開始'
    
    Commands.constants.each do |mod|
      mod_ref = Commands.const_get mod
      if mod_ref.respond_to?(:register_slash_command)
        mod_ref.register_slash_command(BOT)
      end
    end

    puts '[Bot] コマンドの登録完了'
  end

  def self.run
    load_modules
    BOT.run
  rescue StandardError => e
    puts "Bot実行中にエラーが発生しました: #{e.message}"
    puts e.backtrace.join("\n")
  end
end
